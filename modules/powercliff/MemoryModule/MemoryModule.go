package memorymodule

/*
#include "MemoryModule.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)


type Module unsafe.Pointer

func LoadLibrary(data []byte) Module {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	return Module(C.MemoryLoadLibrary(p, C.size_t(len(data))))
}

func FreeLibrary(module Module) {
	C.MemoryFreeLibrary(C.HMEMORYMODULE(module))
}

func GetProcAddress(module Module, name string) uintptr {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	addr := C.MemoryGetProcAddress(C.HMEMORYMODULE(module), cname)
	return uintptr(unsafe.Pointer(addr))
}
