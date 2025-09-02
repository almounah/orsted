package clr

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type SLIST_ENTRY struct {
	Next *SLIST_ENTRY
}
type PSLIST_ENTRY *SLIST_ENTRY

type MemAllocEntry struct {
	AllocEntry      SLIST_ENTRY
	Address         uintptr
	Size            uintptr
	MemAllocTracker uint32
}

type IHostMalloc struct {
	Vtbl         *IHostMallocVtbl
	Count        uint32
	HHeap        uintptr
	MemAllocList *MemAllocEntry
}

type IHostMallocVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Alloc          uintptr
	DebugAlloc     uintptr
	Free           uintptr
}

func IHostMalloc_QueryInterface(this *IHostMalloc, vTableGUID *windows.GUID, ppv **uintptr) uintptr {
	if !IsEqualIID(vTableGUID, &IID_IUnknown) || !IsEqualIID(vTableGUID, &IID_IHostControl) {
		*ppv = nil
		return uintptr(E_NOINTERFACE)
	}
	*ppv = (*uintptr)(unsafe.Pointer(this))

	// Calling AddRef
	syscall.SyscallN(this.Vtbl.AddRef, uintptr(unsafe.Pointer(this)))
	return uintptr(S_OK)
}

func IHostMalloc_AddRef(this *IHostMalloc) uintptr {
	this.Count++
	return uintptr(this.Count)
}

func IHostMalloc_Release(this *IHostMalloc) uintptr {
	this.Count--
	if this.Count == 0 {
		GlobalFree(uintptr(unsafe.Pointer(this)))
	}
	return uintptr(this.Count)
}

func IHostMalloc_Alloc(this *IHostMalloc, cbSize uintptr, eCriticalLevel uint32, ppMem **uintptr) uintptr {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	procHeapAlloc := kernel32.NewProc("HeapAlloc")
	ptr, _, _ := procHeapAlloc.Call(this.HHeap, 0, cbSize)
	*ppMem = (*uintptr)(unsafe.Pointer(ptr))
	if *ppMem == nil {
		return 0x8007000E
	}

	return S_OK
}

func IHostMalloc_DebugAlloc(this *IHostMalloc, cbSize uintptr, eCriticalLevel uint32, psvFileName *byte, iLineNo int, ppMem **uintptr) uintptr {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	procHeapAlloc := kernel32.NewProc("HeapAlloc")
	ptr, _, _ := procHeapAlloc.Call(this.HHeap, 0, cbSize)
	*ppMem = (*uintptr)(unsafe.Pointer(ptr))
	if *ppMem == nil {
		return 0x8007000E
	}

	return S_OK
}

func IHostMalloc_Free(this *IHostMalloc, pMem uintptr) uintptr {
	kernel32 := windows.NewLazyDLL("kernel32.dll")
	procHeapAlloc := kernel32.NewProc("HeapFree")
	procHeapAlloc.Call(this.HHeap, 0, pMem)
	return S_OK
}

func GetHostMalloc() *IHostMalloc {
    var res IHostMalloc

    var resVtbl IHostMallocVtbl
    resVtbl.QueryInterface = windows.NewCallback(IHostMalloc_QueryInterface)
    resVtbl.AddRef = windows.NewCallback(IHostMalloc_AddRef)
    resVtbl.Release = windows.NewCallback(IHostMalloc_Release)
    resVtbl.Alloc = windows.NewCallback(IHostMalloc_Alloc)
    resVtbl.DebugAlloc = windows.NewCallback(IHostMalloc_DebugAlloc)
    resVtbl.Free = windows.NewCallback(IHostMalloc_Free)

    res.Vtbl = &resVtbl

    return &res

}
