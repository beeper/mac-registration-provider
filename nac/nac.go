package nac

// TODO Should this use -fobjc-arc to enable automatic reference counting instead of NSAutoreleasePool?

//#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -Wno-incompatible-pointer-types
//#cgo LDFLAGS: -framework Foundation -framework IOKit
//#include "nac.h"
//#include <dlfcn.h>
import "C"
import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"
)

const identityservicesd = "/System/Library/PrivateFrameworks/IDS.framework/identityservicesd.app/Contents/MacOS/identityservicesd"

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

var ErrNoOffsets = errors.New("no offsets")

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
		return fmt.Errorf("%w for %x", ErrNoOffsets, hash[:])
	}

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
