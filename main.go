package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	_ = json.NewEncoder(os.Stderr).Encode(&versions)

	defer meowMemory()()
	err := nacSanityCheck()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	cert, err := fetchCert(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to fetch cert: %w", err))
	}
	validationCtx, request, err := nacInit(cert)
	if err != nil {
		panic(err)
	}
	sessionInfo, err := fetchInitializeValidation(ctx, request)
	if err != nil {
		panic(fmt.Errorf("failed to initialize validation: %w", err))
	}
	err = nacKeyEstablishment(validationCtx, sessionInfo)
	if err != nil {
		panic(err)
	}
	validationData, err := nacSign(validationCtx)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(validationData))
}
