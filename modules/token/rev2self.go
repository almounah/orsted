package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

//func rev2self() ([]byte, error) {
//	err := windows.RevertToSelf()
//	if err != nil {
//		results := err.Error()
//		return []byte(results), err
//	}
//
//	results := "Successfully reverted to self and dropped the impersonation token"
//	return []byte(results), nil
//}

func rev2self() ([]byte, error) {
	tokenHandle := uintptr(0)
	err := NtSetInformationThread(windows.CurrentThread(), 5, unsafe.Pointer(&tokenHandle), uint32(unsafe.Sizeof(tokenHandle)))
	if err != nil {
		results := err.Error()
		return []byte(results), err
	}

	results := "Successfully reverted to self and dropped the impersonation token"
	return []byte(results), nil
}

