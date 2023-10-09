package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/beeper/nacserv-native/nac"
	"github.com/beeper/nacserv-native/requests"
	"github.com/beeper/nacserv-native/versions"
)

func main() {
	_ = json.NewEncoder(os.Stderr).Encode(&versions.Current)

	err := initSanityCheck()
	if err != nil {
		panic(err)
	}
	err = initFetchCert(context.Background())
	if err != nil {
		panic(err)
	}
	validationData, err := generateValidationData()
	if err != nil {
		panic(err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(validationData))
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

func generateValidationData() ([]byte, error) {
	defer nac.MeowMemory()()

	ctx := context.Background()
	validationCtx, request, err := nac.Init(globalCert)
	if err != nil {
		return nil, err
	}
	sessionInfo, err := requests.InitializeValidation(ctx, request)
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
