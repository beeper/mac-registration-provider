package nac

import (
	"encoding/hex"
	"fmt"
)

var offsets_10_13_6 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "newLocalDeliveryServiceStatString",
	ReferenceAddress:           0x233c34,
	NACInitAddress:             0x3ac270,
	NACKeyEstablishmentAddress: 0x3ac290,
	NACSignAddress:             0x3ac2b0,
}}

var offsets_10_14_6 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "newLocalDeliveryServiceStatString",
	ReferenceAddress:           0x238842,
	NACInitAddress:             0x338ce0,
	NACKeyEstablishmentAddress: 0x3478a0,
	NACSignAddress:             0x333240,
}}

var offsets_10_15_1 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x92787,
	NACInitAddress:             0x3a59e0,
	NACKeyEstablishmentAddress: 0x3b45a0,
	NACSignAddress:             0x39ff40,
}}

var offsets_10_15_2 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x92cb7,
	NACInitAddress:             0x3a67d0,
	NACKeyEstablishmentAddress: 0x3b5390,
	NACSignAddress:             0x3a0d30,
}}

var offsets_10_15_3 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x92c87,
	NACInitAddress:             0x3a67d0,
	NACKeyEstablishmentAddress: 0x3b5390,
	NACSignAddress:             0x3a0d30,
}}

var offsets_10_15_4 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x926cb,
	NACInitAddress:             0x3a5070,
	NACKeyEstablishmentAddress: 0x3b3c30,
	NACSignAddress:             0x39f5d0,
}}

var offsets_10_15_5 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x926bb,
	NACInitAddress:             0x3a5070,
	NACKeyEstablishmentAddress: 0x3b3c30,
	NACSignAddress:             0x39f5d0,
}}

// Offsets support macOS 10.15.6 - 10.15.7 binary
var offsets_10_15_7 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0x9222b,
	NACInitAddress:             0x3a4f70,
	NACKeyEstablishmentAddress: 0x3b3b30,
	NACSignAddress:             0x39f4d0,
}}

// Offsets from the macOS 11.7.7 binary for x86, works on 11.5 - 11.7
var offsets_11_7_7 = imdOffsetTuple{x86: imdOffsets{
	ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
	ReferenceAddress:           0xa3b8e,
	NACInitAddress:             0x3d4870,
	NACKeyEstablishmentAddress: 0x427390,
	NACSignAddress:             0x3c71a0,
}}

// Offsets support macOS 12.7.1 - 12.7.2 binary
var offsets_12_7_2 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xb2278,
		NACInitAddress:             0x4132e0,
		NACKeyEstablishmentAddress: 0x465e00,
		NACSignAddress:             0x405c10,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0x0b562c,
		NACInitAddress:             0x43d408,
		NACKeyEstablishmentAddress: 0x3fdafc,
		NACSignAddress:             0x3f2844,
	},
}

var offsets_13_3_1 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xccfdf,
		NACInitAddress:             0x4ac060,
		NACKeyEstablishmentAddress: 0x48c0a0,
		NACSignAddress:             0x49f390,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xb7570,
		NACInitAddress:             0x414e28,
		NACKeyEstablishmentAddress: 0x40268c,
		NACSignAddress:             0x3dc898,
	},
}

// Offsets from the macOS 13.5 binary, works on 13.5 - 13.6
var offsets_13_6 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xcc743,
		NACInitAddress:             0x4b91e0,
		NACKeyEstablishmentAddress: 0x499220,
		NACSignAddress:             0x4ac510,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xb524c,
		NACInitAddress:             0x41d714,
		NACKeyEstablishmentAddress: 0x40af78,
		NACSignAddress:             0x3e5184,
	},
}

var offsets_14_0 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xd5a4d,
		NACInitAddress:             0x543210,
		NACKeyEstablishmentAddress: 0x523250,
		NACSignAddress:             0x536540,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xc00ec,
		NACInitAddress:             0x4af610,
		NACKeyEstablishmentAddress: 0x49ce74,
		NACSignAddress:             0x477080,
	},
}

var offsets_14_1 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xd6c39,
		NACInitAddress:             0x549b30,
		NACKeyEstablishmentAddress: 0x529b70,
		NACSignAddress:             0x53ce60,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xbf178,
		NACInitAddress:             0x4b2e84,
		NACKeyEstablishmentAddress: 0x4a06e8,
		NACSignAddress:             0x47a8f4,
	},
}

var offsets_14_2 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xd4899,
		NACInitAddress:             0x54c730,
		NACKeyEstablishmentAddress: 0x52c770,
		NACSignAddress:             0x53fa60,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0xbd9f0,
		NACInitAddress:             0x4b55a0,
		NACKeyEstablishmentAddress: 0x4a2e04,
		NACSignAddress:             0x47d010,
	},
}

var offsets_14_3 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0x0d47c9,
		NACInitAddress:             0x54c6d0,
		NACKeyEstablishmentAddress: 0x52c710,
		NACSignAddress:             0x53fa00,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0x0bd81c,
		NACInitAddress:             0x4b5580,
		NACKeyEstablishmentAddress: 0x4a2de4,
		NACSignAddress:             0x47cff0,
	},
}

var offsets_14_4_1 = imdOffsetTuple{
	x86: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0x0d6715,
		NACInitAddress:             0x557cd0,
		NACKeyEstablishmentAddress: 0x537d10,
		NACSignAddress:             0x54b000,
	},
	arm64: imdOffsets{
		ReferenceSymbol:            "IDSProtoKeyTransparencyTrustedServiceReadFrom",
		ReferenceAddress:           0x0c0b84,
		NACInitAddress:             0x4c2468,
		NACKeyEstablishmentAddress: 0x4afccc,
		NACSignAddress:             0x489ed8,
	},
}

// offsets is a map from sha256 hash of identityservicesd to the function pointer offsets in that binary.
var offsets = map[[32]byte]imdOffsetTuple{
	// macOS 10.13.6
	hexToByte32("0d9430e530bfb1eb528152e6f3d062408867bd159d54333228742dd7020312a8"): offsets_10_13_6,
	// macOS 10.14.6
	hexToByte32("23f14e11c672c07ef5934614ae2b83b34065ffe179e4a9bcdcdf00c2b724b3df"): offsets_10_14_6,
	// macOS 10.15.1
	hexToByte32("6423c719735caff7a62ca6ea30da479fa4eb2a8c83255c1340dfcfe5450da2e1"): offsets_10_15_1,
	// macOS 10.15.2
	hexToByte32("30bd65178c67bb8680b967dde7ac636b524ecb870590f8e6ba9af0d898f8d466"): offsets_10_15_2,
	// macOS 10.15.3
	hexToByte32("0031e8fe5e19941c8ce20da12e2abdca61a54b8f8d7e168f83855cca34a44cfd"): offsets_10_15_3,
	// macOS 10.15.4
	hexToByte32("68b96d1beab35116452d33d6fb212b9e23a2795cfe3c91a79148c86f94c7c13e"): offsets_10_15_4,
	// macOS 10.15.5
	hexToByte32("651b8032c0775f0af779f31dee5985dc7d7de56f6732a35069916d5ccde4eaa1"): offsets_10_15_5,
	// macOS 10.15.6 - 10.15.7
	hexToByte32("ff443057a320436216eaf7f5d825ea37b6d4dc05d088a59eac1bf35172eb73b6"): offsets_10_15_7,
	// macOS 11.5.1
	hexToByte32("e9ae1e7f0ef671269bc0b5f3e6791472665c7d17f8e3a3aead6276d15589cd4f"): offsets_11_7_7,
	// macOS 11.6.1
	hexToByte32("f3467734b116f78c22cbe43217d7a337d3cf4dbbc58c0dde81f90dfa19d22e91"): offsets_11_7_7,
	// macOS 11.7.7
	hexToByte32("80107d249088d9762ec38c8f86d6797b5070d476377e7c5ddacf83ad32d00a1e"): offsets_11_7_7,
	// macOS 12.6.3
	hexToByte32("6e8caf477c2b4d3a56a91835a2b6455f36fb0feb13006def7516ac09578c67d0"): {},
	// macOS 12.7.1
	hexToByte32("5833338da6350266eda33f5501c5dfc793e0632b52883aa2389c438c02d03718"): offsets_12_7_2,
	// macOS 12.7.2
	hexToByte32("01aaa511c5d32c5766256a40b5ae8f42fb49b74074dce5936f315244236f15a0"): offsets_12_7_2,
	// macOS 13.2.1
	hexToByte32("4d96de9438fdea5b0b7121e485541ecf0a74489eeb330c151a7d44d289dd3a85"): {},
	// macOS 13.3.1
	hexToByte32("3c8357aaa1df1eb3a21d88182a1a0fca1c612a4d63592e022ca65bbf47deee35"): offsets_13_3_1,
	// macOS 13.5 - 13.6
	hexToByte32("fff8db27fef2a2b874f7bc6fb303a98e3e3b8aceb8dd4c5bfa2bad7b76ea438a"): offsets_13_6,
	// macOS 13.6.3
	hexToByte32("2c674438d30bf489695f2d1b8520afc30cbfb183af82d2fc53d74ce39a25b24e"): offsets_13_6,
	// macOS 13.6.4
	hexToByte32("8f22dcfda56a4d3c38931f20fe33db1a6720e4d8571e452aa5a8b56b4c69842a"): offsets_13_6,
	// macOS 14.0
	hexToByte32("9ffda11206ef874b1e6cb1d8f8fed330d2ac2cbbc87afc15485f4e4371afcd9a"): offsets_14_0,
	// macOS 14.1 - 14.1.2
	hexToByte32("2483dc690217e959d386ae4573bacb8d669f3c0a666b1874ebfcb8131a9c18d7"): offsets_14_1,
	// macOS 14.1.2 (M3 Only)
	hexToByte32("47aa51e63ced0bb00dd27dab0def6f065a1a4911e250b79761681865fbd03644"): offsets_14_1,
	// macOS 14.2
	hexToByte32("034fc179e1cce559931a8e46866f54154cb1c5413902319473537527a2702b64"): offsets_14_2,
	// macOS 14.3
	hexToByte32("d3c6986fefcbd2efea2a8a7c88104bf22d60d1f4f2bbf3615a1e3ce098aba765"): offsets_14_3,
	// macOS 14.4.1
	hexToByte32("b82c5c6c9010a42cb64397e3760dd31144cbd471126111de9bb27fa3d2d2639a"): offsets_14_4_1,
}

type imdOffsetTuple struct {
	x86   imdOffsets
	arm64 imdOffsets
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
