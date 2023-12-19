package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/beeper/mac-registration-provider/versions"
)

type WebsocketRequest[T any] struct {
	Command string `json:"command"`
	ReqID   int    `json:"id,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type RegisterBody struct {
	Code   string `json:"code,omitempty"`
	Secret string `json:"secret,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

type EmptyResponse struct{}

type VersionsResponse struct {
	Versions versions.Versions `json:"versions"`
}

type ValidationDataResponse struct {
	Data       []byte    `json:"data"`
	ValidUntil time.Time `json:"valid_until"`
}

var dataCache ValidationDataResponse
var cacheLock sync.Mutex

func cachedGenerateData(ctx context.Context) (ValidationDataResponse, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	if time.Now().UTC().Add(5 * time.Minute).After(dataCache.ValidUntil) {
		data, validUntil, err := GenerateValidationData(ctx)
		if err != nil {
			return ValidationDataResponse{}, err
		}
		dataCache = ValidationDataResponse{Data: data, ValidUntil: validUntil}
	}
	return dataCache, nil
}

func handleCommand(ctx context.Context, req WebsocketRequest[json.RawMessage]) (any, error) {
	switch req.Command {
	case "pong":
		return nil, nil
	case "ping":
		// Pre-cache validation data on ping
		go func() {
			_, err := cachedGenerateData(ctx)
			if err != nil {
				log.Printf("Failed to pregenerate validation data on ping: %v", err)
			} else {
				log.Println("Pregenerated validation data on ping")
			}
		}()
		return EmptyResponse{}, nil
	case "get-version-info":
		return VersionsResponse{Versions: versions.Current}, nil
	case "get-validation-data":
		return cachedGenerateData(ctx)
	default:
		return nil, fmt.Errorf("unknown command %q", req.Command)
	}
}

type RelayConfig struct {
	Code   string `json:"code"`
	Secret string `json:"secret"`
}

func readConfig() (string, *RelayConfig, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "beeper-validation-provider", "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config RelayConfig
	if configData != nil {
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return "", nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	return configPath, &config, nil
}

func writeConfig(cfg *RelayConfig, configPath string) error {
	err := os.MkdirAll(filepath.Dir(configPath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(cfg)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func ConnectRelay(ctx context.Context, addr string) error {
	configPath, config, err := readConfig()
	if err != nil {
		return err
	}

	c, _, err := websocket.Dial(ctx, addr+"/api/v1/provider", &websocket.DialOptions{
		HTTPHeader: http.Header{
			"User-Agent": []string{submitUserAgent},
		},
	})
	if err != nil {
		return err
	}
	defer c.CloseNow()

	err = wsjson.Write(ctx, c, &WebsocketRequest[*RegisterBody]{
		Command: "register",
		ReqID:   1,
		Data: &RegisterBody{
			Code:   config.Code,
			Secret: config.Secret,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to write register request: %w", err)
	}

	var registerResp WebsocketRequest[*RegisterBody]
	err = wsjson.Read(ctx, c, &registerResp)
	if err != nil {
		return fmt.Errorf("failed to read register response: %w", err)
	} else if registerResp.Command != "response" || registerResp.ReqID != 1 {
		return fmt.Errorf("unexpected register response %+v", registerResp)
	} else if registerResp.Data.Error != "" {
		_ = os.Rename(configPath, configPath+".bak")
		return fmt.Errorf("failed to register: %s", registerResp.Data.Error)
	}

	if config.Code == "" || config.Code != registerResp.Data.Code {
		if config.Code != "" {
			log.Println("Registration token changed")
		}
		config.Code = registerResp.Data.Code
		config.Secret = registerResp.Data.Secret
		err = writeConfig(config, configPath)
		if err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}
	}

	if *jsonOutput {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{
			"code": registerResp.Data.Code,
			"path": configPath,
		})
	} else {
		fmt.Println()
		fmt.Println(" ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
		fmt.Println(" ┃ iMessage registration code:", registerResp.Data.Code, "┃")
		fmt.Println(" ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
		fmt.Println()
		fmt.Println("Delete", configPath, "if you want to regenerate the token")
	}

	reqID := 1
	cancelableCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				reqID++
				err = wsjson.Write(ctx, c, WebsocketRequest[any]{
					Command: "ping",
					ReqID:   reqID,
				})
			case <-cancelableCtx.Done():
			}
		}
	}()

	log.Printf("Connection successful")
	for {
		var req WebsocketRequest[json.RawMessage]
		err = wsjson.Read(ctx, c, &req)
		if err != nil {
			return fmt.Errorf("failed to read request: %w", err)
		}
		log.Printf("Received command %s/%d", req.Command, req.ReqID)
		resp, err := handleCommand(ctx, req)
		if err != nil {
			log.Printf("Command %s/%d failed: %v", req.Command, req.ReqID, err)
			resp = ErrorResponse{Error: err.Error()}
		} else if resp == nil {
			continue
		} else {
			log.Printf("Command %s/%d succeeded", req.Command, req.ReqID)
		}
		err = wsjson.Write(ctx, c, WebsocketRequest[any]{
			Command: "response",
			ReqID:   req.ReqID,
			Data:    resp,
		})
		if err != nil {
			return fmt.Errorf("failed to write response to %d: %w", req.ReqID, err)
		}
	}
}
