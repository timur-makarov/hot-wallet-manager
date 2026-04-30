package types

// #include <TrustWalletCore/TWString.h>
import "C"

import (
	"unsafe"
)

// TWStringGoString converts a C.TWString* into a Go string.
func TWStringGoString(s unsafe.Pointer) string {
	return C.GoString(C.TWStringUTF8Bytes(s))
}

// TWStringCreateWithGoString converts a Go string into a C.TWString* (caller owns the result).
func TWStringCreateWithGoString(s string) unsafe.Pointer {
	cStr := C.CString(s)
	defer C.free(unsafe.Pointer(cStr))
	return C.TWStringCreateWithUTF8Bytes(cStr)
}
