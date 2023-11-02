package nac

import (
	"encoding/hex"
	"fmt"
)

// offsets is a map from sha256 hash of identityservicesd to the function pointer offsets in that binary.
var offsets = map[[32]byte]imdOffsets{
	// macOS 11.5.1 (maybe any 11.5.x)
	hexToByte32("e9ae1e7f0ef671269bc0b5f3e6791472665c7d17f8e3a3aead6276d15589cd4f"): {},
	// macOS 11.6.1 (maybe any 11.6.x)
	hexToByte32("f3467734b116f78c22cbe43217d7a337d3cf4dbbc58c0dde81f90dfa19d22e91"): {},
	// macOS 11.7.7 (maybe any 11.7.x)
	hexToByte32("80107d249088d9762ec38c8f86d6797b5070d476377e7c5ddacf83ad32d00a1e"): {},
	// macOS 13.5 - 13.6 (possibly earlier versions too)
	hexToByte32("fff8db27fef2a2b874f7bc6fb303a98e3e3b8aceb8dd4c5bfa2bad7b76ea438a"): {
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xb524c,
		NACInitAddress:             0x41d714,
		NACKeyEstablishmentAddress: 0x40af78,
		NACSignAddress:             0x3e5184,
	},
	// macOS 14.0
	hexToByte32("9ffda11206ef874b1e6cb1d8f8fed330d2ac2cbbc87afc15485f4e4371afcd9a"): {
		// TODO
	},
	// macOS 14.1
	hexToByte32("2483dc690217e959d386ae4573bacb8d669f3c0a666b1874ebfcb8131a9c18d7"): {
		// TODO
	},
}

type imdOffsets struct {
	ReferenceSymbol            string
	ReferenceAddress           int
	NACInitAddress             int
	NACKeyEstablishmentAddress int
	NACSignAddress             int
}

func hexToByte32(val string) [32]byte {
	out, err := hex.DecodeString(val)
	if err != nil {
		panic(err)
	} else if len(out) != 32 {
		panic(fmt.Errorf("expected 32 bytes, got %d", len(out)))
	}
	return *(*[32]byte)(out)
}
