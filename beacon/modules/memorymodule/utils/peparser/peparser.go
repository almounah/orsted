package peparser

import (
	"errors"
	"unsafe"

	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
)

const (
	IMAGE_DIRECTORY_ENTRY_EXPORT    = 0
	IMAGE_DIRECTORY_ENTRY_IMPORT    = 1
	IMAGE_DIRECTORY_ENTRY_EXCEPTION = 3
	IMAGE_DIRECTORY_ENTRY_BASERELOC = 5
	IMAGE_DIRECTORY_ENTRY_TLS       = 9

	IMAGE_FILE_DLL = 0x2000
)

func InitializePeStruct(dllByte []byte) (memmoduletypes.PPE_HDRS, error) {
	var peHdrs memmoduletypes.PE_HDRS
	dwPeSize := len(dllByte)

	dllPointer := uintptr(unsafe.Pointer(&dllByte[0]))

	// Populating Data
	peHdrs.DwFileSize = memmoduletypes.DWORD(dwPeSize)
	peHdrs.PFileBuffer = memmoduletypes.PBYTE(unsafe.Pointer(&dllByte[0]))

	// Getting Img NT Header from e_lfanew
	e_lfanew := *((*uint32)(unsafe.Pointer(dllPointer + 0x3c)))

	peHdrs.PImgNtHdrs = memmoduletypes.PIMAGE_NT_HEADERS64(unsafe.Pointer(dllPointer + uintptr(e_lfanew)))

	// Checking Signature cause why not
	if peHdrs.PImgNtHdrs.Signature != memmoduletypes.IMAGE_NT_SIGNATURE {
		return nil, errors.New("Not a Valid File")
	}

	// Getting the first section address by translating addresses
	var sectionAddr uintptr
	sectionAddr = dllPointer + uintptr(e_lfanew) + unsafe.Sizeof(peHdrs.PImgNtHdrs.Signature) + unsafe.Sizeof(peHdrs.PImgNtHdrs.OptionalHeader) + unsafe.Sizeof(peHdrs.PImgNtHdrs.FileHeader)
	peHdrs.PImgeSectionHdrs = (memmoduletypes.PIMAGE_SECTION_HEADER)(unsafe.Pointer(sectionAddr))

	// Populating Directories
	peHdrs.PEntryExportDataDir = &peHdrs.PImgNtHdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXPORT]
	peHdrs.PEntryImportDataDir = &peHdrs.PImgNtHdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_IMPORT]
	peHdrs.PEntryBaseRelocDataDir = &peHdrs.PImgNtHdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_BASERELOC]
	peHdrs.PEntryExceptionDataDir = &peHdrs.PImgNtHdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_EXCEPTION]
	peHdrs.PEntryTLSDataDir = &peHdrs.PImgNtHdrs.OptionalHeader.DataDirectory[IMAGE_DIRECTORY_ENTRY_TLS]

	// Checking if it is a Dll
	if peHdrs.PImgNtHdrs.FileHeader.Characteristics&IMAGE_FILE_DLL != 0 {
		peHdrs.BIsDLLFile = true
	} else {
        peHdrs.BIsDLLFile = false
    }

	return &peHdrs, nil
}
