package data

import "unsafe"

// Str2bytes Fast convert
func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Bytes2str Fast convert
func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
