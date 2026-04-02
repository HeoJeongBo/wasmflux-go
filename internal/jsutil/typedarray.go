//go:build js && wasm

package jsutil

import (
	"math"
	"syscall/js"
	"unsafe"
)

// CopyFloat64sFromJS copies a JS Float64Array into a Go []float64 slice.
// Uses js.CopyBytesToGo for efficient transfer.
func CopyFloat64sFromJS(dst []float64, src js.Value) int {
	n := src.Length()
	if n > len(dst) {
		n = len(dst)
	}
	if n == 0 {
		return 0
	}

	buf := make([]byte, n*8)
	uint8Array := js.Global().Get("Uint8Array").New(src.Get("buffer"), src.Get("byteOffset").Int(), n*8)
	js.CopyBytesToGo(buf, uint8Array)

	for i := 0; i < n; i++ {
		bits := uint64(buf[i*8]) |
			uint64(buf[i*8+1])<<8 |
			uint64(buf[i*8+2])<<16 |
			uint64(buf[i*8+3])<<24 |
			uint64(buf[i*8+4])<<32 |
			uint64(buf[i*8+5])<<40 |
			uint64(buf[i*8+6])<<48 |
			uint64(buf[i*8+7])<<56
		dst[i] = math.Float64frombits(bits)
	}

	return n
}

// CopyFloat64sToJS copies a Go []float64 slice into a JS Float64Array.
func CopyFloat64sToJS(dst js.Value, src []float64) int {
	n := len(src)
	dstLen := dst.Length()
	if n > dstLen {
		n = dstLen
	}
	if n == 0 {
		return 0
	}

	buf := make([]byte, n*8)
	for i := 0; i < n; i++ {
		bits := math.Float64bits(src[i])
		buf[i*8] = byte(bits)
		buf[i*8+1] = byte(bits >> 8)
		buf[i*8+2] = byte(bits >> 16)
		buf[i*8+3] = byte(bits >> 24)
		buf[i*8+4] = byte(bits >> 32)
		buf[i*8+5] = byte(bits >> 40)
		buf[i*8+6] = byte(bits >> 48)
		buf[i*8+7] = byte(bits >> 56)
	}

	uint8Array := js.Global().Get("Uint8Array").New(dst.Get("buffer"), dst.Get("byteOffset").Int(), n*8)
	js.CopyBytesToJS(uint8Array, buf)

	return n
}

// NewFloat64Array creates a new JS Float64Array from a Go slice.
func NewFloat64Array(src []float64) js.Value {
	arr := js.Global().Get("Float64Array").New(len(src))
	CopyFloat64sToJS(arr, src)
	return arr
}

// BytesFromJS copies a JS Uint8Array into a Go []byte slice.
func BytesFromJS(src js.Value) []byte {
	n := src.Length()
	dst := make([]byte, n)
	js.CopyBytesToGo(dst, src)
	return dst
}

// BytesToJS creates a JS Uint8Array from a Go []byte slice.
func BytesToJS(src []byte) js.Value {
	arr := js.Global().Get("Uint8Array").New(len(src))
	js.CopyBytesToJS(arr, src)
	return arr
}

// Sizeof returns the byte size of a type. Used internally for typed array calculations.
func Sizeof[T any]() int {
	var zero T
	return int(unsafe.Sizeof(zero))
}
