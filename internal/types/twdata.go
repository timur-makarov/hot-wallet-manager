package types

// #include <TrustWalletCore/TWData.h>
import "C"

import (
	"encoding/hex"
	"unsafe"
)

// TWDataGoBytes converts a C.TWData* into a Go []byte.
func TWDataGoBytes(d unsafe.Pointer) []byte {
	cBytes := C.TWDataBytes(d)
	cSize := C.TWDataSize(d)
	return C.GoBytes(unsafe.Pointer(cBytes), C.int(cSize))
}

// TWDataCreateWithGoBytes converts a Go []byte into a C.TWData* (caller owns the result).
func TWDataCreateWithGoBytes(d []byte) unsafe.Pointer {
	cBytes := C.CBytes(d)
	defer C.free(unsafe.Pointer(cBytes))
	return C.TWDataCreateWithBytes((*C.uchar)(cBytes), C.size_t(len(d)))
}

// TWDataHexString converts a C.TWData* into a Go hex string.
func TWDataHexString(d unsafe.Pointer) string {
	return hex.EncodeToString(TWDataGoBytes(d))
}
