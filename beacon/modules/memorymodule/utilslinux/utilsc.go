package utils

/*
#include <fcntl.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>
#include <string.h>
#include <errno.h>
#include <dlfcn.h>
*/
import "C"
import (
	"fmt"
	"orsted/beacon/utils"
	"unsafe"
)

func shm_open(name string) (int, error) {
    shmName := C.CString(name)

	fd := C.shm_open(shmName, C.O_RDWR|C.O_CREAT, C.S_IRWXU)
	if fd == -1 {
		utils.Print("shm_open failed")
		return -1, fmt.Errorf("Failed to Open shm")
	}
    return int(fd), nil
}

func dlopen(path string) (uintptr, error) {
    pathName := C.CString(path)
    res := C.dlopen(pathName, C.RTLD_LAZY)
	if res == nil {
		return uintptr(res), fmt.Errorf("Failed to Open with dlopen")
	}
    return uintptr(res), nil
}

func dlsym(module uintptr, symbole string) (uintptr, error) {
    symName := C.CString(symbole)
    res := C.dlsym(unsafe.Pointer(module), symName)
	if res == nil {
		return uintptr(res), fmt.Errorf("Failed to Find symbol")
	}
    return uintptr(res), nil
}

