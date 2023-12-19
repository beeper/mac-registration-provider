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
var jsonOutput = flag.Bool("json", false, "Output JSON instead of text")
var submitUserAgent = fmt.Sprintf("mac-validation-provider/%s go/%s macOS/%s", Commit[:8], strings.TrimPrefix(runtime.Version(), "go"), versions.Current.SoftwareVersion)
var once = flag.Bool("once", false, "Generate a single validation data, print it to stdout and exit")

func main() {
	flag.Parse()
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
				})
			}
			log.Fatalf("No offsets found for %s/%s (hash: %x)", noOffsetsErr.Version, noOffsetsErr.Arch, noOffsetsErr.Hash[:])
			return
		}
		panic(err)
	}
	log.Println("Running sanity check...")
	err = InitSanityCheck()
	if err != nil {
		panic(err)
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
		err = ConnectRelay(context.Background(), *relayServer)
		if err != nil {
			log.Printf("Error in relay connection: %v", err)
		}
	}
}
