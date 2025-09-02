package clr

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type MyMemoryManager struct {
	Vtlb          *MyMemoryManagerVtbl
	Count         uint32
	MallocManager *IHostMalloc
	MemAllocList  *MemAllocEntry
}

type MyMemoryManagerVtbl struct {
	QueryInterface                     uintptr
	AddRef                             uintptr
	Release                            uintptr
	CreateMalloc                       uintptr
	VirtualAlloc                       uintptr
	VirtualFree                        uintptr
	VirtualQuery                       uintptr
	VirtualProtect                     uintptr
	GetMemoryLoad                      uintptr
	RegisterMemoryNotificationCallback uintptr
	NeedsVirtualAddressSpace           uintptr
	AcquiredVirtualAddressSpace        uintptr
	ReleasedVirtualAddressSpace        uintptr
}

func MemoryManager_QueryInterface(this *MyMemoryManager, vTableGUID *windows.GUID, ppv **uintptr) uintptr {
	if !IsEqualIID(vTableGUID, &IID_IUnknown) || !IsEqualIID(vTableGUID, &IID_IHostControl) {
		*ppv = nil
		return uintptr(E_NOINTERFACE)
	}
	*ppv = (*uintptr)(unsafe.Pointer(this))

	// Calling AddRef
	syscall.SyscallN(this.Vtlb.AddRef, uintptr(unsafe.Pointer(this)))
	return uintptr(S_OK)
}

func MemoryManager_AddRef(this *MyMemoryManager) uintptr {
	this.Count++
	return uintptr(this.Count)
}

func MemoryManager_Release(this *MyMemoryManager) uintptr {
	this.Count--
	if this.Count == 0 {
		GlobalFree(uintptr(unsafe.Pointer(this)))
	}
	return uintptr(this.Count)
}

func MemoryManager_CreateMalloc(this *MyMemoryManager, dwMallocType uint32, ppMalloc **IHostMalloc) uintptr {
	mallocManager := GetHostMalloc()

	var hHeap uintptr
	if dwMallocType&0x2 != 0 {
		hHeap = HeapCreate(0x00040000)
	} else {
		hHeap = HeapCreate(0)
	}

	mallocManager.HHeap = hHeap
	mallocManager.MemAllocList = this.MemAllocList

	*ppMalloc = mallocManager
	this.MallocManager = mallocManager

	return S_OK
}

func MemoryManager_VirtualAlloc(this *MyMemoryManager, pAddress uintptr, dwSize uintptr, flAllocationType uint32, flProtect uint32, eCriticalLevel uint32, ppMem **uintptr) uintptr {
	allcAddr, _ := windows.VirtualAlloc(pAddress, dwSize, flAllocationType, flProtect)
	*ppMem = (*uintptr)(unsafe.Pointer(allcAddr))
    fmt.Println("Called VirtualAlloc")
	return S_OK
}

func MemoryManager_VirtualFree(this *MyMemoryManager, lpAddress uintptr, dwSize uintptr, dwFreeType uint32) uintptr {
	windows.VirtualFree(lpAddress, dwSize, dwFreeType)
    fmt.Println("Called VirtualFree")
	return S_OK
}

func MemoryManager_VirtualQuery(this *MyMemoryManager, lpAddress uintptr, lpBuffer *uintptr, dwLength uintptr, pResult *uintptr) uintptr {
	r1, _, _ := syscall.SyscallN(windows.NewLazyDLL("kernel32.dll").NewProc("VirtualQuery").Addr(), lpAddress, uintptr(unsafe.Pointer(lpBuffer)), uintptr(dwLength))
	*pResult = r1
    fmt.Println("Called VirtualQuery")
	return S_OK
}

func MemoryManager_VirtualProtect(this *MyMemoryManager, lpAddress uintptr, dwSize uintptr, flNewProtect uint32, pOldProtect *uint32) uintptr {
	windows.VirtualProtect(lpAddress, dwSize, flNewProtect, pOldProtect)
    fmt.Println("Called VirtualProtect")
	return S_OK
}

func MemoryManager_GetMemoryLoad(this *MyMemoryManager, pMemoryLoad *uint32, pAvailableBytes *uintptr) uintptr {
	*pMemoryLoad = 30
	*pAvailableBytes = 100 * 1024 * 1024
	return S_OK
}

func MemoryManager_RegisterMemoryNotificationCallback(this *MyMemoryManager, pCallBack uintptr) uintptr {
	return S_OK
}

func MemoryManager_NeedsVirtualAddressSpace(this *MyMemoryManager, startAddress *uintptr, size uintptr) uintptr {
	return S_OK
}

func MemoryManager_AcquiredVirtualAddressSpace(this *MyMemoryManager, startAddress uintptr, size uintptr) uintptr {
    allocEntry := &MemAllocEntry{}
    allocEntry.Address = startAddress
    allocEntry.Size = size
    allocEntry.MemAllocTracker = 3

    (PSLIST_ENTRY)(unsafe.Pointer(allocEntry)).Next = (PSLIST_ENTRY)(unsafe.Pointer(this.MemAllocList)).Next
    (PSLIST_ENTRY)(unsafe.Pointer(this.MemAllocList)).Next = (PSLIST_ENTRY)(unsafe.Pointer(allocEntry))

    fmt.Println("Called VirtualProtect")

    return S_OK
}

func MemoryManager_ReleasedVirtualAddressSpace(this *MyMemoryManager, startAddress *uintptr) uintptr {
    return S_OK
}

func GetMemoryManager() *MyMemoryManager{

    memoryManager := &MyMemoryManager{}
    vtable := &MyMemoryManagerVtbl{}
    vtable.QueryInterface = windows.NewCallback(MemoryManager_QueryInterface)
    vtable.AddRef = windows.NewCallback(MemoryManager_AddRef)
    vtable.Release = windows.NewCallback(MemoryManager_Release)
    vtable.VirtualAlloc = windows.NewCallback(MemoryManager_VirtualAlloc)
    vtable.VirtualFree = windows.NewCallback(MemoryManager_VirtualFree)
    vtable.VirtualQuery = windows.NewCallback(MemoryManager_VirtualQuery)
    vtable.NeedsVirtualAddressSpace = windows.NewCallback(MemoryManager_NeedsVirtualAddressSpace)
    vtable.AcquiredVirtualAddressSpace = windows.NewCallback(MemoryManager_AcquiredVirtualAddressSpace)
    vtable.RegisterMemoryNotificationCallback = windows.NewCallback(MemoryManager_RegisterMemoryNotificationCallback)
    vtable.ReleasedVirtualAddressSpace = windows.NewCallback(MemoryManager_ReleasedVirtualAddressSpace)
    vtable.CreateMalloc = windows.NewCallback(MemoryManager_CreateMalloc)

    memoryManager.Vtlb = vtable

    memAllocListHead := &MemAllocEntry{}
    (PSLIST_ENTRY)(unsafe.Pointer(memAllocListHead)).Next = nil

    memoryManager.MemAllocList = (*MemAllocEntry)(unsafe.Pointer(&memAllocListHead))
    return memoryManager
}
