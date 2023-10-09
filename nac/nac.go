package nac

// TODO Should this use -fobjc-arc to enable automatic reference counting instead of NSAutoreleasePool?

//#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -Wno-incompatible-pointer-types
//#cgo LDFLAGS: -framework Foundation -framework IOKit
//#include "nac.h"
//#include <dlfcn.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	IMDPath                        = "/System/Library/PrivateFrameworks/IDS.framework/identityservicesd.app/Contents/MacOS/identityservicesd"
	IMDReferenceSymbol             = "IDSProtoKeyTransparencyTrustedServiceReadFrom"
	IMDReferenceAddress            = 0xb524c
	IMDNACInitAddress              = 0x41d714
	IMDNACKeyEstablishmentAdddress = 0x40af78
	IMDNACSignAddress              = 0x3e5184
)

var base uintptr

func Load() error {
	handle := C.dlopen(C.CString(IMDPath), C.RTLD_LAZY)
	if handle == nil {
		return fmt.Errorf("failed to load %s: %v", IMDPath, C.GoString(C.dlerror()))
	}
	ref := C.dlsym(handle, C.CString(IMDReferenceSymbol))
	if ref == nil {
		return fmt.Errorf("failed to find %s at %x: %v", IMDReferenceSymbol, IMDReferenceAddress, C.GoString(C.dlerror()))
	}
	base = uintptr(ref) - IMDReferenceAddress
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
	resp := int(C.nacInitProxy(unsafe.Pointer(base+IMDNACInitAddress), nil, C.int(0), nil, nil, nil))
	if resp != -44023 {
		return fmt.Errorf("NACInit sanity check had unexpected response %d", resp)
	}
	return nil
}

func Init(cert []byte) (validationCtx unsafe.Pointer, request []byte, err error) {
	var outputBytesLen C.int
	var outputBytesPtr unsafe.Pointer
	resp := int(C.nacInitProxy(
		unsafe.Pointer(base+IMDNACInitAddress),
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
		unsafe.Pointer(base+IMDNACKeyEstablishmentAdddress),
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
		unsafe.Pointer(base+IMDNACSignAddress),
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
