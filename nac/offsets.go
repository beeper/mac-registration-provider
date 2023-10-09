package nac

import (
	"encoding/hex"
	"fmt"
)

// offsets is a map from sha256 hash of identityservicesd to the function pointer offsets in that binary.
var offsets = map[[32]byte]imdOffsets{
	hexToByte32("fff8db27fef2a2b874f7bc6fb303a98e3e3b8aceb8dd4c5bfa2bad7b76ea438a"): {
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xb524c,
		NACInitAddress:             0x41d714,
		NACKeyEstablishmentAddress: 0x40af78,
		NACSignAddress:             0x3e5184,
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
