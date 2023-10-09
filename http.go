package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"howett.net/plist"
)

var client = &http.Client{}

const (
	validationCertURL       = "http://static.ess.apple.com/identity/validation/cert-1.0.plist"
	initializeValidationURL = "https://identity.ess.apple.com/WebObjects/TDIdentityService.woa/wa/initializeValidation"
)

type CertResponse struct {
	Cert []byte `plist:"cert"`
}

func makeRequest(ctx context.Context, url string, body, output any) error {
	method := http.MethodGet
	var bodyReader io.Reader
	if body != nil {
		method = http.MethodPost
		var buf bytes.Buffer
		err := plist.NewEncoder(&buf).Encode(body)
		if err != nil {
			return fmt.Errorf("failed to encode body: %w", err)
		}
		bodyReader = &buf
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to prepare request: %w", err)
	}
	req.Header.Set("User-Agent", versions.UserAgent())
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/x-apple-plist")
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}
	_, err = plist.Unmarshal(respData, output)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	return nil
}

func fetchCert(ctx context.Context) ([]byte, error) {
	var parsedResp CertResponse
	err := makeRequest(ctx, validationCertURL, nil, &parsedResp)
	return parsedResp.Cert, err
}

type ReqInitializeValidation struct {
	SessionInfoRequest []byte `plist:"session-info-request"`
}

type RespInitializeValidation struct {
	SessionInfo []byte `plist:"session-info"`
}

func fetchInitializeValidation(ctx context.Context, request []byte) ([]byte, error) {
	var parsedResp RespInitializeValidation
	err := makeRequest(ctx, initializeValidationURL, &ReqInitializeValidation{request}, &parsedResp)
	return parsedResp.SessionInfo, err
}
