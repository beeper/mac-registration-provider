package nac

// TODO Should this use -fobjc-arc to enable automatic reference counting?

//#cgo CFLAGS: -x objective-c -Wno-deprecated-declarations -Wno-incompatible-pointer-types
//#cgo LDFLAGS: -framework Foundation -framework IOKit
//#include "nac.h"
//#include "meowMemory.h"
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"
)

func MeowMemory() func() {
	runtime.LockOSThread()
	pool := C.meowMakePool()
	return func() {
		C.meowReleasePool(pool)
		runtime.UnlockOSThread()
	}
}

func SanityCheck() error {
	resp := int(C.NACInit(nil, C.int(0), nil, nil, nil))
	if resp != -44023 {
		return fmt.Errorf("NACInit sanity check had unexpected response %d", resp)
	}
	return nil
}

func Init(cert []byte) (validationCtx unsafe.Pointer, request []byte, err error) {
	var outputBytesLen C.int
	var outputBytesPtr unsafe.Pointer
	resp := int(C.NACInit(
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
	resp := int(C.NACKeyEstablishment(
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
	resp := int(C.NACSign(
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
