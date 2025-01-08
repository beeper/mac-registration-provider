package nac

// TODO Should this use -fobjc-arc to enable automatic reference counting instead of NSAutoreleasePool?

//#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -Wno-incompatible-pointer-types
//#cgo LDFLAGS: -framework Foundation -framework IOKit
//#include "nac.h"
//#include <dlfcn.h>
import "C"
import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"

	"github.com/beeper/mac-registration-provider/find_offsets"
	"github.com/beeper/mac-registration-provider/versions"
)

const identityservicesd = "/System/Library/PrivateFrameworks/IDS.framework/identityservicesd.app/Contents/MacOS/identityservicesd"
const symbol = "IDSProtoKeyTransparencyTrustedServiceReadFrom"

var nacInitAddr, nacKeyEstablishmentAddr, nacSignAddr unsafe.Pointer

func sha256sum(path string) (hash [32]byte, err error) {
	hasher := sha256.New()
	var file *os.File
	if file, err = os.Open(path); err != nil {
		err = fmt.Errorf("failed to open %q: %w", path, err)
	} else if _, err = io.Copy(hasher, file); err != nil {
		err = fmt.Errorf("failed to hash %q: %w", path, err)
	} else {
		hash = *(*[32]byte)(hasher.Sum(nil))
	}
	return
}

type NoOffsetsError struct {
	Hash    string `json:"hash"`
	Version string `json:"version"`
	BuildID string `json:"build_id"`
	Arch    string `json:"arch"`
}

func (err NoOffsetsError) Error() string {
	return fmt.Sprintf("no offsets for %s/%s/%s (hash: %s)", err.Version, err.BuildID, err.Arch, err.Hash)
}

func Load() error {
	hash, err := sha256sum(identityservicesd)
	if err != nil {
		return err
	}
	var offs imdOffsets
	if runtime.GOARCH == "arm64" {
		offs = offsets[hash].arm64
	} else {
		offs = offsets[hash].x86
	}
	if offs.ReferenceSymbol == "" {
		// Call the FindOffsets function directly
		newOffsets, err := FindOffsets(identityservicesd)
		if err != nil {
			return fmt.Errorf("failed to find offsets: %v", err)
		}

		if runtime.GOARCH == "arm64" {
			offs = newOffsets.arm64
		} else {
			offs = newOffsets.x86
		}

		if offs.ReferenceSymbol == "" {
			return NoOffsetsError{
				Hash:    hex.EncodeToString(hash[:]),
				Version: versions.Current.SoftwareVersion,
				BuildID: versions.Current.SoftwareBuildID,
				Arch:    runtime.GOARCH,
			}
		}
	}

	fmt.Printf("Reference Symbol: %s\n", offs.ReferenceSymbol)
	fmt.Printf("Reference Address: %06x\n", offs.ReferenceAddress)
	fmt.Printf("NAC Init Address: %06x\n", offs.NACInitAddress)
	fmt.Printf("NAC Key Establishment Address: %06x\n", offs.NACKeyEstablishmentAddress)
	fmt.Printf("NAC Sign Address: %06x\n", offs.NACSignAddress)

	handle := C.dlopen(C.CString(identityservicesd), C.RTLD_LAZY)
	if handle == nil {
		return fmt.Errorf("failed to load %s: %v", identityservicesd, C.GoString(C.dlerror()))
	}
	ref := C.dlsym(handle, C.CString(offs.ReferenceSymbol))
	if ref == nil {
		return fmt.Errorf("failed to find %s at %x: %v", offs.ReferenceSymbol, offs.ReferenceAddress, C.GoString(C.dlerror()))
	}
	base := unsafe.Add(unsafe.Pointer(ref), -offs.ReferenceAddress)
	nacInitAddr = unsafe.Add(base, offs.NACInitAddress)
	nacKeyEstablishmentAddr = unsafe.Add(base, offs.NACKeyEstablishmentAddress)
	nacSignAddr = unsafe.Add(base, offs.NACSignAddress)
	return nil
}

func FindOffsets(filePath string) (imdOffsetTuple, error) {
	architectures, err := find_offsets.ScanMachOFATBinary(filePath)
	if err != nil {
		return imdOffsetTuple{}, err
	}

	searchResults := find_offsets.SearchInArchitectures(filePath, architectures, find_offsets.HexStringsModern)
	offsets := imdOffsetTuple{
		x86: imdOffsets{
			ReferenceSymbol:            symbol,
			ReferenceAddress:           searchResults[0]["ReferenceAddress (_IDSProtoKeyTransparencyTrustedServiceReadFrom)"][0],
			NACInitAddress:             searchResults[0]["NACInitAddress"][0],
			NACKeyEstablishmentAddress: searchResults[0]["NACKeyEstablishmentAddress"][0],
			NACSignAddress:             searchResults[0]["NACSignAddress"][0],
		},
		arm64: imdOffsets{
			ReferenceSymbol:            symbol,
			ReferenceAddress:           searchResults[1]["ReferenceAddress (_IDSProtoKeyTransparencyTrustedServiceReadFrom)"][0],
			NACInitAddress:             searchResults[1]["NACInitAddress"][0],
			NACKeyEstablishmentAddress: searchResults[1]["NACKeyEstablishmentAddress"][0],
			NACSignAddress:             searchResults[1]["NACSignAddress"][0],
		},
	}

	return offsets, nil
}

func MeowMemory() func() {
	runtime.LockOSThread()
	pool := C.meowMakePool()
	return func() {
		C.meowReleasePool(pool)
		runtime.UnlockOSThread()
	}
}

func SanityCheck() error {
	resp := int(C.nacInitProxy(nacInitAddr, nil, C.int(0), nil, nil, nil))
	if resp != -44023 {
		return fmt.Errorf("NACInit sanity check had unexpected response %d", resp)
	}
	return nil
}

func Init(cert []byte) (validationCtx unsafe.Pointer, request []byte, err error) {
	var outputBytesLen C.int
	var outputBytesPtr unsafe.Pointer
	resp := int(C.nacInitProxy(
		nacInitAddr,
		unsafe.Pointer(&cert[0]),
		C.int(len(cert)),
		&validationCtx,
		&outputBytesPtr,
		&outputBytesLen,
	))
	if resp != 0 {
		err = fmt.Errorf("NACInit failed with response %d", resp)
		return
	}
	request = unsafe.Slice((*byte)(outputBytesPtr), int(outputBytesLen))
	return
}

func KeyEstablishment(validationCtx unsafe.Pointer, response []byte) (err error) {
	resp := int(C.nacKeyEstablishmentProxy(
		nacKeyEstablishmentAddr,
		validationCtx,
		unsafe.Pointer(&response[0]),
		C.int(len(response)),
	))
	if resp != 0 {
		err = fmt.Errorf("NACKeyEstablishment failed with response %d", resp)
		return
	}
	return
}

func Sign(validationCtx unsafe.Pointer) (validationData []byte, err error) {
	var outputBytesPtr unsafe.Pointer
	var outputBytesLen C.int
	resp := int(C.nacSignProxy(
		nacSignAddr,
		validationCtx,
		nil,
		C.int(0),
		&outputBytesPtr,
		&outputBytesLen,
	))
	if resp != 0 {
		err = fmt.Errorf("NACSign failed with response %d", resp)
		return
	}
	validationData = unsafe.Slice((*byte)(outputBytesPtr), int(outputBytesLen))
	return
}
