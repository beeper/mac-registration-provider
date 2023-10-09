package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strings"
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

var submitURL = flag.String("url", "", "URL to submit validation data to")
var submitToken = flag.String("token", "", "Token to include when submitting validation data")
var submitInterval = flag.Duration("interval", 5*time.Minute, "Interval at which to submit new validation data to the server")
var submitUserAgent = fmt.Sprintf("nacserv-native/%s go/%s macOS/%s", Version, strings.TrimPrefix(runtime.Version(), "go"), versions.Current.SoftwareVersion)
var once = flag.Bool("once", false, "Generate a single validation data, print it to stdout and exit")

func main() {
	flag.Parse()
	if !*once {
		parsedURL, err := url.Parse(*submitURL)
		if err != nil {
			panic(fmt.Errorf("failed to parse input URL: %w", err))
		} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
			panic(fmt.Errorf("unexpected URL scheme %q", parsedURL.Scheme))
		}
	} else {
		log.SetOutput(io.Discard)
	}

	log.Println("Running sanity check...")
	err := initSanityCheck()
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
		fmt.Println(base64.StdEncoding.EncodeToString(validationData))
		return
	}
	for {
		log.Println("Generating validation data...")
		if validationData, err := generateValidationData(context.Background()); err != nil {
			log.Printf("Failed to generate validation data: %v", err)
		} else if err = submitValidationData(context.Background(), validationData); err != nil {
			log.Printf("Failed to submit validation data: %v", err)
		} else {
			log.Println("Successfully generated and submitted validation data")
		}
		time.Sleep(*submitInterval)
	}
}

func submitValidationData(ctx context.Context, data []byte) error {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&ReqSubmitValidationData{ValidationData: data, DeviceInfo: versions.Current})
	if err != nil {
		return fmt.Errorf("failed to encode request payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, *submitURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("User-Agent", submitUserAgent)
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
