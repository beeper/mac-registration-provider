package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/beeper/nacserv-native/nac"
	"github.com/beeper/nacserv-native/requests"
	"github.com/beeper/nacserv-native/versions"
)

type ReqSubmitValidationData struct {
	ValidationData []byte            `json:"validation_data"`
	DeviceInfo     versions.Versions `json:"device_info"`
}

const Version = "0.1.0"

var submitToken = flag.String("token", "", "Token to include when submitting validation data")
var submitInterval = flag.Duration("interval", 3*time.Minute, "Interval at which to submit new validation data to the server")
var submitUserAgent = fmt.Sprintf("nacserv-native/%s go/%s macOS/%s", Version, strings.TrimPrefix(runtime.Version(), "go"), versions.Current.SoftwareVersion)
var once = flag.Bool("once", false, "Generate a single validation data, print it to stdout and exit")

func main() {
	flag.Parse()
	var urls []string
	if !*once {
		urls = flag.Args()
		if len(urls) == 0 {
			_, _ = fmt.Fprintln(os.Stderr, "You must pass one or more URLs to submit to when not using -once")
			return
		}
		for _, u := range urls {
			parsedURL, err := url.Parse(u)
			if err != nil {
				panic(fmt.Errorf("failed to parse input URL %q: %w", u, err))
			} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
				panic(fmt.Errorf("unexpected URL scheme %q", parsedURL.Scheme))
			}
		}
	} else {
		log.SetOutput(io.Discard)
	}

	log.Println("Loading identityservicesd")
	err := nac.Load()
	if err != nil {
		panic(err)
	}
	log.Println("Running sanity check...")
	err = initSanityCheck()
	if err != nil {
		panic(err)
	}
	log.Println("Fetching certificate...")
	err = initFetchCert(context.Background())
	if err != nil {
		panic(err)
	}
	log.Println("Initialization complete")
	if *once {
		validationData, err := generateValidationData(context.Background())
		if err != nil {
			panic(err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(&ReqSubmitValidationData{
			ValidationData: validationData,
			DeviceInfo:     versions.Current,
		})
		return
	}
	for {
		log.Println("Generating validation data...")
		if validationData, err := generateValidationData(context.Background()); err != nil {
			log.Printf("Failed to generate validation data: %v", err)
		} else {
			submitValidationDataToURLs(context.Background(), urls, validationData)
		}
		time.Sleep(*submitInterval)
	}
}

func submitValidationDataToURLs(ctx context.Context, urls []string, data []byte) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, u := range urls {
		go func(addr string) {
			defer wg.Done()
			err := submitValidationData(ctx, addr, data)
			if err != nil {
				log.Printf("Failed to submit validation data to %s: %v", addr, err)
			} else {
				log.Println("Submitted validation data to", addr)
			}
		}(u)
	}
	wg.Wait()
}

func submitValidationData(ctx context.Context, url string, data []byte) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&ReqSubmitValidationData{ValidationData: data, DeviceInfo: versions.Current})
	if err != nil {
		return fmt.Errorf("failed to encode request payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("User-Agent", submitUserAgent)
	req.Header.Set("Content-Type", "application/json")
	if len(*submitToken) > 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *submitToken))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}
	return nil
}

var globalCert []byte

func initFetchCert(ctx context.Context) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var err error
	globalCert, err = requests.FetchCert(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch cert: %w", err)
	}
	return nil
}

func initSanityCheck() error {
	defer nac.MeowMemory()()
	return nac.SanityCheck()
}

func generateValidationData(ctx context.Context) ([]byte, error) {
	defer nac.MeowMemory()()

	validationCtx, request, err := nac.Init(globalCert)
	if err != nil {
		return nil, err
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	sessionInfo, err := requests.InitializeValidation(ctx, request)
	cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize validation: %w", err)
	}
	err = nac.KeyEstablishment(validationCtx, sessionInfo)
	if err != nil {
		return nil, err
	}
	validationData, err := nac.Sign(validationCtx)
	if err != nil {
		return nil, err
	}
	return validationData, nil
}
