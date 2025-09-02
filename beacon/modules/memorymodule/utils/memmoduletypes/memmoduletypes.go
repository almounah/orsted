package memmoduletypes

import (
	"golang.org/x/sys/windows"
)

const (
	IMAGE_DIRECTORY_ENTRY_BASERELOC         = 5
	IMAGE_REL_BASED_ABSOLUTE                = 0
	IMAGE_REL_BASED_HIGH                    = 1
	IMAGE_REL_BASED_LOW                     = 2
	IMAGE_REL_BASED_HIGHLOW                 = 3
	IMAGE_REL_BASED_DIR64                   = 10
	IMAGE_SCN_CNT_INITIALIZED_DATA          = 0x00000040
	IMAGE_SCN_CNT_UNINITIALIZED_DATA        = 0x00000080
	IMAGE_SCN_MEM_DISCARDABLE        uint32 = 0x02000000
	IMAGE_SCN_MEM_EXECUTE                   = 0x20000000
	IMAGE_SCN_MEM_READ                      = 0x40000000
	IMAGE_SCN_MEM_WRITE                     = 0x80000000
	IMAGE_SCN_MEM_NOT_CACHED                = 0x04000000

	PAGE_NOACCESS          = windows.PAGE_NOACCESS
	PAGE_WRITECOPY         = windows.PAGE_WRITECOPY
	PAGE_READONLY          = windows.PAGE_READONLY
	PAGE_READWRITE         = windows.PAGE_READWRITE
	PAGE_EXECUTE           = windows.PAGE_EXECUTE
	PAGE_EXECUTE_WRITECOPY = windows.PAGE_EXECUTE_WRITECOPY
	PAGE_EXECUTE_READ      = windows.PAGE_EXECUTE_READ
	PAGE_EXECUTE_READWRITE = windows.PAGE_EXECUTE_READWRITE
	PAGE_NOCACHE           = windows.PAGE_NOCACHE

	IMAGE_NT_SIGNATURE = 0x00004550
)

type (
	DWORD     uint32
	WORD      uint16
	BYTE      uint8
	PBYTE     *BYTE
	ULONGLONG uint64
)

type IMAGE_NT_HEADERS64 struct {
	Signature      DWORD
	FileHeader     IMAGE_FILE_HEADER
	OptionalHeader IMAGE_OPTIONAL_HEADER64
}
type PIMAGE_NT_HEADERS64 *IMAGE_NT_HEADERS64

// IMAGE_OPTIONAL_HEADER64 represents the optional header for 64-bit architecture.
type IMAGE_OPTIONAL_HEADER64 struct {
	Magic                       WORD
	MajorLinkerVersion          BYTE
	MinorLinkerVersion          BYTE
	SizeOfCode                  DWORD
	SizeOfInitializedData       DWORD
	SizeOfUninitializedData     DWORD
	AddressOfEntryPoint         DWORD
	BaseOfCode                  DWORD
	ImageBase                   ULONGLONG
	SectionAlignment            DWORD
	FileAlignment               DWORD
	MajorOperatingSystemVersion WORD
	MinorOperatingSystemVersion WORD
	MajorImageVersion           WORD
	MinorImageVersion           WORD
	MajorSubsystemVersion       WORD
	MinorSubsystemVersion       WORD
	Win32VersionValue           DWORD
	SizeOfImage                 DWORD
	SizeOfHeaders               DWORD
	CheckSum                    DWORD
	Subsystem                   WORD
	DllCharacteristics          WORD
	SizeOfStackReserve          ULONGLONG
	SizeOfStackCommit           ULONGLONG
	SizeOfHeapReserve           ULONGLONG
	SizeOfHeapCommit            ULONGLONG
	LoaderFlags                 DWORD
	NumberOfRvaAndSizes         DWORD

	DataDirectory [16]IMAGE_DATA_DIRECTORY
}
type PIMAGE_OPTIONAL_HEADER64 *IMAGE_OPTIONAL_HEADER64

// IMAGE_DATA_DIRECTORY represents a data directory entry.
type IMAGE_DATA_DIRECTORY struct {
	VirtualAddress DWORD
	Size           DWORD
}
type PIMAGE_DATA_DIRECTORY *IMAGE_DATA_DIRECTORY

// IMAGE_FILE_HEADER represents the file header in the IMAGE_NT_HEADERS structure.
type IMAGE_FILE_HEADER struct {
	Machine              WORD
	NumberOfSections     WORD
	TimeDateStamp        DWORD
	PointerToSymbolTable DWORD
	NumberOfSymbols      DWORD
	SizeOfOptionalHeader WORD
	Characteristics      WORD
}
type PIMAGE_FILE_HEADER *IMAGE_FILE_HEADER

type IMAGE_SECTION_HEADER struct {
	Name                 [8]byte
	VirtualSize          DWORD
	VirtualAddress       DWORD
	SizeOfRawData        DWORD
	PointerToRawData     DWORD
	PointerToRelocations DWORD
	PointerToLinenumbers DWORD
	NumberOfRelocations  WORD
	NumberOfLinenumbers  WORD
	Characteristics      DWORD
}
type PIMAGE_SECTION_HEADER *IMAGE_SECTION_HEADER

type IMAGE_BASE_RELOCATION struct {
	VirtualAddress DWORD
	BlockSize      DWORD
}
type PIMAGE_BASE_RELOCATION *IMAGE_BASE_RELOCATION

// BASE_RELOCATION_ENTRY represents the base relocation entry structure
type BASE_RELOCATION_ENTRY struct {
	OffsetType WORD // Combined field for Offset and Type
}
type PBASE_RELOCATION_ENTRY *BASE_RELOCATION_ENTRY

// Offset extracts the Offset from the combined field
func (bre BASE_RELOCATION_ENTRY) Offset() WORD {
	return bre.OffsetType & 0xFFF
}

// Type extracts the Type from the combined field
func (bre BASE_RELOCATION_ENTRY) Type() WORD {
	return (bre.OffsetType >> 12) & 0xF
}

type IMAGE_IMPORT_DESCRIPTOR struct {
	Characteristics DWORD
	TimeDateStamp   DWORD
	ForwarderChain  DWORD
	Name            DWORD
	FirstThunk      DWORD
}

type IMAGE_IMPORT_BY_NAME struct {
	Hint uint16
	Name [1]byte
}

type IMAGE_THUNK_DATA struct {
	AddressOfData uint64
}

type IMAGE_TLS_DIRECTORY64 struct {
	StartAddressOfRawData ULONGLONG
	EndAddressOfRawData   ULONGLONG
	AddressOfIndex        ULONGLONG
	AddressOfCallBacks    ULONGLONG
	SizeOfZeroFill        DWORD
	Reserved0             DWORD
	Alignment             DWORD
	Reserved1             DWORD
}
type PIMAGE_TLS_DIRECTORY64 *IMAGE_TLS_DIRECTORY64

type PE_HDRS struct {
	PFileBuffer            PBYTE
	DwFileSize             DWORD
	PImgNtHdrs             PIMAGE_NT_HEADERS64
	PImgeSectionHdrs       PIMAGE_SECTION_HEADER
	PEntryImportDataDir    PIMAGE_DATA_DIRECTORY
	PEntryBaseRelocDataDir PIMAGE_DATA_DIRECTORY
	PEntryTLSDataDir       PIMAGE_DATA_DIRECTORY
	PEntryExceptionDataDir PIMAGE_DATA_DIRECTORY
	PEntryExportDataDir    PIMAGE_DATA_DIRECTORY
	BIsDLLFile             bool
}
type PPE_HDRS *PE_HDRS
