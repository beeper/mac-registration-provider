package main

import (
	"context"
	"fmt"
	"time"

	"github.com/beeper/nacserv-native/nac"
	"github.com/beeper/nacserv-native/requests"
)

var globalCert []byte

func InitFetchCert(ctx context.Context) error {
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

func InitSanityCheck() error {
	defer nac.MeowMemory()()
	return nac.SanityCheck()
}

const ValidityTime = 15 * time.Minute

func GenerateValidationData(ctx context.Context) ([]byte, time.Time, error) {
	defer nac.MeowMemory()()

	validationCtx, request, err := nac.Init(globalCert)
	if err != nil {
		return nil, time.Time{}, err
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	// Record valid until time before request, so it's definitely valid for at least that long
	validUntil := time.Now().UTC().Add(ValidityTime)
	sessionInfo, err := requests.InitializeValidation(ctx, request)
	cancel()
	if err != nil {
		return nil, validUntil, fmt.Errorf("failed to initialize validation: %w", err)
	}
	err = nac.KeyEstablishment(validationCtx, sessionInfo)
	if err != nil {
		return nil, validUntil, err
	}
	validationData, err := nac.Sign(validationCtx)
	if err != nil {
		return nil, validUntil, err
	}
	return validationData, validUntil, nil
}
