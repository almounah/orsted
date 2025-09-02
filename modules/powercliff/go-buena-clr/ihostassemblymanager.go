package clr

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type MyAssemblyManager struct {
	PVtbl          *MyAssemblyManager_Vtbl
	AssemblyStore  *MyAssemblyStore
	TargAssemb *TargetAssembly
	Count          uint32
}

type MyAssemblyManager_Vtbl struct {
	QueryInterface            uintptr
	AddRef                    uintptr
	Release                   uintptr
	GetNonHostStoreAssemblies uintptr
	GetAssemblyStore          uintptr
}

func MyAssemblyManager_QueryInterface(this *MyHostControl, vTableGUID *windows.GUID, ppv **uintptr) uintptr {
    if !IsEqualIID(vTableGUID, &IID_IUnknown) || !IsEqualIID(vTableGUID, &IID_IHostControl) {
        *ppv = nil
        return uintptr(E_NOINTERFACE)
    }
    *ppv = (*uintptr)(unsafe.Pointer(this))

    // Calling AddRef
    syscall.SyscallN(this.PVtbl.AddRef, uintptr(unsafe.Pointer(this)))
    return uintptr(S_OK)

}

func MyAssemblyManager_AddRef(this *MyHostControl) uintptr {
    this.Count++
    return uintptr(this.Count)
}

func MyAssemblyManager_Release(this *MyHostControl) uintptr {
    this.Count--
    if this.Count == 0 {
        GlobalFree(uintptr(unsafe.Pointer(this)))
    }
    return uintptr(this.Count)
}

func MyAssemblyManager_GetNonHostStoreAssemblies(this *MyAssemblyManager, ppReferenceList **ICLRAssemblyReferenceList ) uintptr {
    *ppReferenceList = nil
    return S_OK
}

func MyAssemblyManager_GetAssemblyStore(this *MyAssemblyManager, ppAssemblyStore **MyAssemblyStore ) uintptr {
    assemblyStore := GetNewCustomAssemblyStore()
    assemblyStore.tgAssembly = this.TargAssemb
    this.AssemblyStore = &assemblyStore
    *ppAssemblyStore = this.AssemblyStore

    return S_OK
}

func GetNewCustomAssemblyManager() MyAssemblyManager {

    var res MyAssemblyManager
    var vtable MyAssemblyManager_Vtbl

    vtable.QueryInterface = windows.NewCallback(MyAssemblyManager_QueryInterface)
    vtable.AddRef = windows.NewCallback(MyAssemblyManager_AddRef)
    vtable.Release = windows.NewCallback(MyAssemblyManager_Release)
    vtable.GetNonHostStoreAssemblies = windows.NewCallback(MyAssemblyManager_GetNonHostStoreAssemblies)
    vtable.GetAssemblyStore = windows.NewCallback(MyAssemblyManager_GetAssemblyStore)

    res.PVtbl = &vtable
    return res
}
