package helper

import (
	"unsafe"

)


func UintptrToBytes(ptr uintptr) []byte {
	// Create a pointer to the uintptr value
	ptrPtr := unsafe.Pointer(&ptr)

	// Convert the pointer to a byte slice
	byteSlice := make([]byte, unsafe.Sizeof(ptr))
	for i := 0; i < int(unsafe.Sizeof(ptr)); i++ {
		byteSlice[i] = *(*byte)(unsafe.Pointer(uintptr(ptrPtr) + uintptr(i)))
	}

	return byteSlice
}
