package clr

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type IHostAssemblyStore struct {
	PVtbl          *MyAssemblyStore_Vtbl
}

type MyAssemblyStore struct {
	PVtbl          *MyAssemblyStore_Vtbl
    tgAssembly *TargetAssembly
	Count          uint32
}

type MyAssemblyStore_Vtbl struct {
	QueryInterface  uintptr
	AddRef          uintptr
	Release         uintptr
	ProvideAssembly uintptr
	ProvideModule   uintptr
}

type AssemblyBindInfo struct {
	DwAppDomainId        uint32
	LpReferencedIdentity *uint16
	LpPostPolicyIdentity *uint16
	EPolicyLevel         uint32
}

type ModuleBindInfo struct {
	DwAppDomainId        uint32
	LpReferencedIdentity *uint16
	LpModuleName *uint16
}
func MyAssemblyStore_QueryInterface(this *MyHostControl, vTableGUID *windows.GUID, ppv **uintptr) uintptr {
    if !IsEqualIID(vTableGUID, &IID_IUnknown) || !IsEqualIID(vTableGUID, &IID_IHostControl) {
        *ppv = nil
        return uintptr(E_NOINTERFACE)
    }
    *ppv = (*uintptr)(unsafe.Pointer(this))

    // Calling AddRef
    syscall.SyscallN(this.PVtbl.AddRef, uintptr(unsafe.Pointer(this)))
    return uintptr(S_OK)

}

func MyAssemblyStore_AddRef(this *MyHostControl) uintptr {
    this.Count++
    return uintptr(this.Count)
}

func MyAssemblyStore_Release(this *MyHostControl) uintptr {
    this.Count--
    if this.Count == 0 {
        GlobalFree(uintptr(unsafe.Pointer(this)))
    }
    return uintptr(this.Count)
}

func MyAssemblyStore_ProvideAssembly(this *MyAssemblyStore, pBindInfo *AssemblyBindInfo, pAssemblyId *uint64, pContext *uint64, ppStmAssemblyImage **IStream, ppStmPDB **IStream) uintptr {
    str1 := windows.UTF16PtrToString(this.tgAssembly.AssemblyInfo)
    str2 := windows.UTF16PtrToString(pBindInfo.LpPostPolicyIdentity)
    if str1 == str2 {
        *pContext = 0
        *pAssemblyId = 50000

        assemblyStream := SHCreateMemStream(this.tgAssembly.AssemblyBytes, this.tgAssembly.AssemblySize)
        *ppStmAssemblyImage = assemblyStream
        
        return S_OK
    }
    return uintptr(HResultFromWin32(2))
}
func MyAssemblyStore_ProvideModule(this *MyAssemblyStore, pBindInfo *ModuleBindInfo, pAssemblyId *uint64, pContext *uint64, ppStmAssemblyImage **IStream, ppStmPDB **IStream) uintptr {
    return uintptr(HResultFromWin32(2))
}


func GetNewCustomAssemblyStore() MyAssemblyStore{

    var res MyAssemblyStore
    var vtable MyAssemblyStore_Vtbl

    vtable.QueryInterface = windows.NewCallback(MyAssemblyStore_QueryInterface)
    vtable.AddRef = windows.NewCallback(MyAssemblyStore_AddRef)
    vtable.Release = windows.NewCallback(MyAssemblyStore_Release)
    vtable.ProvideAssembly = windows.NewCallback(MyAssemblyStore_ProvideAssembly)
    vtable.ProvideModule = windows.NewCallback(MyAssemblyStore_ProvideModule)

    res.PVtbl = &vtable
    return res
}
