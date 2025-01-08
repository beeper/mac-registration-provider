package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/beeper/mac-registration-provider/find_offsets"
	"github.com/beeper/mac-registration-provider/nac"
	"github.com/beeper/mac-registration-provider/versions"
)

type ReqSubmitValidationData struct {
	ValidationData []byte            `json:"validation_data"`
	ValidUntil     time.Time         `json:"valid_until"`
	NacservCommit  string            `json:"nacserv_commit"`
	DeviceInfo     versions.Versions `json:"device_info"`
}

var Commit = "unknown "

var submitToken = flag.String("submit-token", "", "Token to include when submitting validation data")
var submitInterval = flag.Duration("submit-interval", 0, "Interval at which to submit new validation data to the server")
var relayServer = flag.String("relay-server", "https://registration-relay.beeper.com", "URL of the relay server to use")
var overrideConfigPath = flag.String("config-path", "", "File to save registration code in when using relay mode")
var jsonOutput = flag.Bool("json", false, "Output JSON instead of text")
var submitUserAgent = fmt.Sprintf("mac-registration-provider/%s go/%s macOS/%s", Commit[:8], strings.TrimPrefix(runtime.Version(), "go"), versions.Current.SoftwareVersion)
var once = flag.Bool("once", false, "Generate a single validation data, print it to stdout and exit")
var checkCompatibility = flag.Bool("check-compatibility", false, "Check if offsets for the current OS version are available and exit")
var shouldFindOffsets = flag.Bool("find-offsets", false, "Find offsets in the specified binary")
var identityServiceDPath = flag.String("identityservicesd", "/System/Library/PrivateFrameworks/IDS.framework/identityservicesd.app/Contents/MacOS/identityservicesd", "Path to the identityservicesd binary")

func main() {
	flag.Parse()

	if *shouldFindOffsets {
		find_offsets.PrintOffsets(*identityServiceDPath)
		return
	}

	var urls []string
	if *submitInterval > 0 {
		urls = flag.Args()
		if len(urls) == 0 {
			_, _ = fmt.Fprintln(os.Stderr, "You must pass one or more URLs to submit to when using -interval")
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
	}

	log.Printf("Starting mac-registration-provider %s", Commit[:8])
	log.Println("Loading identityservicesd")
	err := nac.Load()
	if err != nil {
		var noOffsetsErr nac.NoOffsetsError
		if errors.As(err, &noOffsetsErr) {
			if *jsonOutput {
				_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
					"error": "no offsets",
					"data":  err,
					"ok":    false,
				})
			}
			log.Fatalf("No offsets found for %s/%s/%s (hash: %s)", noOffsetsErr.Version, noOffsetsErr.BuildID, noOffsetsErr.Arch, noOffsetsErr.Hash)
			return
		}
		panic(err)
	}
	log.Println("Running sanity check...")
	safetyExitCancel := make(chan struct{})
	go func() {
		select {
		case <-time.After(5 * time.Second):
			log.Fatalln("Sanity check timed out")
		case <-safetyExitCancel:
		}
	}()
	err = InitSanityCheck()
	if err != nil {
		panic(err)
	}
	close(safetyExitCancel)
	if *checkCompatibility {
		log.Println("Compatibility check successful")
		if *jsonOutput {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
				"ok": true,
			})
		}
		return
	}
	log.Println("Fetching certificate...")
	err = InitFetchCert(context.Background())
	if err != nil {
		panic(err)
	}
	log.Println("Initialization complete")
	if *once {
		validationData, validUntil, err := GenerateValidationData(context.Background())
		if err != nil {
			panic(err)
		}
		_ = json.NewEncoder(os.Stdout).Encode(&ReqSubmitValidationData{
			ValidationData: validationData,
			ValidUntil:     validUntil,
			NacservCommit:  Commit,
			DeviceInfo:     versions.Current,
		})
		return
	}
	if *submitInterval > 0 {
		log.Printf("Submit mode: periodically submitting validation data to %+v", urls)
		for {
			generateAndSubmit(urls)
		}
	} else if len(*relayServer) > 0 {
		log.Printf("Relay mode: responding to requests over websocket at %s", *relayServer)
		reconnectIn := 2 * time.Second
		lastReconnect := time.Now()
		for {
			err = ConnectRelay(context.Background(), *relayServer)
			if err == nil {
				break
			} else if strings.HasPrefix(err.Error(), "failed to register:") {
				log.Printf("Error in relay connection: %v, not reconnecting", err)
				if *jsonOutput {
					_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
						"error": "registration rejected",
					})
				}
				os.Exit(10)
			}
			log.Printf("Error in relay connection: %v, reconnecting in %v", err, reconnectIn)
			time.Sleep(reconnectIn)
			if time.Since(lastReconnect) < 5*time.Minute {
				if reconnectIn < 1*time.Minute {
					reconnectIn *= 2
				}
			} else {
				reconnectIn = 2 * time.Second
			}
		}
	}
}
