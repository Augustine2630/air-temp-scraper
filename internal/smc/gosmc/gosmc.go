//go:build darwin

// Package gosmc provides low-level CGo bindings to Apple's IOKit SMC interface.
// Adapted from https://github.com/dkorunic/iSMC (GPLv2+, Copyright 2019 Dinko Korunic).
package gosmc

// #cgo CFLAGS: -O2 -Wall
// #cgo LDFLAGS: -framework IOKit
// #include <stdlib.h>
// #include "smc.h"
import "C"

import "unsafe"

// IOReturn success/error codes (subset used by temperature reading).
const (
	IOReturnSuccess = 0x0
)

// SMC type string constants.
const (
	TypeFLT  = "flt"
	TypeUI8  = "ui8"
	TypeUI16 = "ui16"
	TypeUI32 = "ui32"
	TypeFLAG = "flag"
)

// UInt32Char is a 5-byte null-terminated SMC key or type identifier.
type UInt32Char [5]byte

// ToString returns the UInt32Char as a Go string.
func (bs UInt32Char) ToString() string { return string(bs[:]) }

func (bs UInt32Char) toC() C.UInt32Char_t {
	var xs C.UInt32Char_t
	for i := range bs {
		xs[i] = C.char(bs[i])
	}
	return xs
}

func uint32CharFromC(xs C.UInt32Char_t) UInt32Char {
	var bs UInt32Char
	for i := range xs {
		bs[i] = byte(xs[i])
	}
	return bs
}

// SMCBytes holds the raw data bytes returned by an SMC key read.
type SMCBytes [32]byte

func smcBytesFromC(xs C.SMCBytes_t) SMCBytes {
	var bs SMCBytes
	for i := range xs {
		bs[i] = byte(xs[i])
	}
	return bs
}

// SMCVal is the value returned by SMCReadKey.
type SMCVal struct {
	Key      UInt32Char
	DataSize uint32
	DataType UInt32Char
	Bytes    SMCBytes
}

// SMCOpen opens a connection to the named IOKit service (e.g. "AppleSMC").
func SMCOpen(service string) (connection uint, result int) {
	svc := C.CString(service)
	defer C.free(unsafe.Pointer(svc))

	var conn C.uint
	result = int(C.SMCOpen(svc, &conn))
	return uint(conn), result
}

// SMCClose releases the IOKit connection handle.
func SMCClose(connection uint) int {
	return int(C.SMCClose(C.uint(connection)))
}

// SMCReadKey reads the value of the given 4-char SMC key.
func SMCReadKey(connection uint, key string) (*SMCVal, int) {
	k := C.CString(key)
	defer C.free(unsafe.Pointer(k))

	v := C.SMCVal_t{}
	result := C.SMCReadKey(C.uint(connection), k, &v)

	return &SMCVal{
		Key:      uint32CharFromC(v.key),
		DataSize: uint32(v.dataSize),
		DataType: uint32CharFromC(v.dataType),
		Bytes:    smcBytesFromC(v.bytes),
	}, int(result)
}
