package sections

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
	"orsted/beacon/utils"
)

const (
	IMAGE_SCN_MEM_EXECUTE    = 0x20000000
	IMAGE_SCN_MEM_READ       = 0x40000000
	IMAGE_SCN_MEM_WRITE      = 0x80000000
	IMAGE_SCN_MEM_NOT_CACHED = 0x04000000
)

func ChangePermission(pBaseAddress uintptr, pNtHdrs memmoduletypes.PIMAGE_NT_HEADERS64, pImgFirstSection memmoduletypes.PIMAGE_SECTION_HEADER) error {
	firstSectionAddr := uintptr(unsafe.Pointer(pImgFirstSection))
	for i := 0; i < int(pNtHdrs.FileHeader.NumberOfSections); i++ {

		sectionAddr := firstSectionAddr + uintptr(i)*unsafe.Sizeof(memmoduletypes.IMAGE_SECTION_HEADER{})
		section := (memmoduletypes.PIMAGE_SECTION_HEADER)(unsafe.Pointer(sectionAddr))

        if section.VirtualAddress == 0 || section.SizeOfRawData == 0 {
            continue
        }

		sectionDestination := pBaseAddress + uintptr(section.VirtualAddress)

		var dwProtection uint32

		charac := section.Characteristics

		if charac&IMAGE_SCN_MEM_WRITE != 0 {
			dwProtection = windows.PAGE_WRITECOPY
		}

		if charac&IMAGE_SCN_MEM_READ != 0 {
			dwProtection = windows.PAGE_READONLY
		}

		if charac&IMAGE_SCN_MEM_WRITE != 0 && charac&IMAGE_SCN_MEM_READ != 0 {
			dwProtection = windows.PAGE_READWRITE
		}

		if charac&IMAGE_SCN_MEM_EXECUTE != 0 {
			dwProtection = windows.PAGE_EXECUTE
		}

		if charac&IMAGE_SCN_MEM_EXECUTE != 0 && charac&IMAGE_SCN_MEM_READ != 0 {
			dwProtection = windows.PAGE_EXECUTE_READ
		}

		if charac&IMAGE_SCN_MEM_EXECUTE != 0 && charac&IMAGE_SCN_MEM_WRITE != 0 {
			dwProtection = windows.PAGE_EXECUTE_WRITECOPY
		}

		if charac&IMAGE_SCN_MEM_EXECUTE != 0 && charac&IMAGE_SCN_MEM_WRITE != 0 && charac&IMAGE_SCN_MEM_READ != 0 {
			dwProtection = windows.PAGE_EXECUTE_READWRITE
		}

		var oldprotect uint32

		err := windows.VirtualProtect(sectionDestination, uintptr(section.SizeOfRawData), dwProtection, &oldprotect)
		if err != nil {
			utils.Print("Error in changing section permission", err.Error())
			return err
		}

	}
	return nil
}
