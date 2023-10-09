package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/beeper/nacserv-native/nac"
	"github.com/beeper/nacserv-native/requests"
	"github.com/beeper/nacserv-native/versions"
)

type ReqSubmitValidationData struct {
	ValidationData []byte            `json:"validation_data"`
	DeviceInfo     versions.Versions `json:"device_info"`
}

var submitURL = flag.String("url", "", "URL to submit validation data to")
var submitInterval = flag.Duration("duration", 5*time.Minute, "Interval at which to submit new validation data to the server")

func main() {
	flag.Parse()
	parsedURL, err := url.Parse(*submitURL)
	if err != nil {
		panic(fmt.Errorf("failed to parse input URL: %w", err))
	} else if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		panic(fmt.Errorf("unexpected URL scheme %q", parsedURL.Scheme))
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
