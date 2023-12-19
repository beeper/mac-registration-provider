package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/beeper/mac-registration-provider/versions"
)

var panicCounter = 0

func generateAndSubmit(urls []string) {
	defer func() {
		err := recover()
		if err != nil {
			panicCounter++
			log.Printf("Panic while generating validation data: %v\n%s", err, debug.Stack())
			sleepDuration := time.Duration(panicCounter) * 5 * time.Minute
			log.Println("Sleeping for", sleepDuration)
			time.Sleep(sleepDuration)
		}
	}()
	log.Println("Generating validation data...")
	if validationData, validUntil, err := GenerateValidationData(context.Background()); err != nil {
		log.Printf("Failed to generate validation data: %v", err)
	} else {
		submitValidationDataToURLs(context.Background(), urls, validationData, validUntil)
	}
	panicCounter = 0
	time.Sleep(*submitInterval)
}

func submitValidationDataToURLs(ctx context.Context, urls []string, data []byte, validUntil time.Time) {
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(len(urls))
	for _, u := range urls {
		go func(addr string) {
			defer wg.Done()
			err := submitValidationData(ctx, addr, data, validUntil)
			if err != nil {
				log.Printf("Failed to submit validation data to %s: %v", addr, err)
			} else {
				log.Println("Submitted validation data to", addr)
			}
		}(u)
	}
	wg.Wait()
}

func submitValidationData(ctx context.Context, url string, data []byte, validUntil time.Time) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&ReqSubmitValidationData{
		ValidationData: data,
		ValidUntil:     validUntil,
		NacservCommit:  Commit,
		DeviceInfo:     versions.Current,
	})
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
