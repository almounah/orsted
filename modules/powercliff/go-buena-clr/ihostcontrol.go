package clr

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type MyHostControl struct {
	PVtbl        *MyHostControl_Vtbl
	TargAssembly *TargetAssembly
	MemoryManager *MyMemoryManager
	Count        uint32
}

type IHostControl struct {
	PVtbl *MyHostControl_Vtbl
	Count uint32
}

type MyHostControl_Vtbl struct {
	QueryInterface      uintptr
	AddRef              uintptr
	Release             uintptr
	GetHostManager      uintptr
	SetAppDomainManager uintptr
}

func MyQueryInterface(this *MyHostControl, vTableGUID *windows.GUID, ppv **uintptr) uintptr {
	if !IsEqualIID(vTableGUID, &IID_IUnknown) || !IsEqualIID(vTableGUID, &IID_IHostControl) {
		*ppv = nil
		return uintptr(E_NOINTERFACE)
	}
	*ppv = (*uintptr)(unsafe.Pointer(this))

	// Calling AddRef
	syscall.SyscallN(this.PVtbl.AddRef, uintptr(unsafe.Pointer(this)))
	return uintptr(S_OK)
}

func MyAddRef(this *MyHostControl) uintptr {
	this.Count++
	return uintptr(this.Count)
}

func MyRelease(this *MyHostControl) uintptr {
	this.Count--
	if this.Count == 0 {
		GlobalFree(uintptr(unsafe.Pointer(this)))
	}
	return uintptr(this.Count)
}

func MySetAppDomainManager(this *MyHostControl, dwAppDomainId uint32, pUnkAppDomainManager *IUnknown) uintptr {
	return uintptr(E_NOTIMPL)
}

func MyGetHostManager(this *MyHostControl, riid *windows.GUID, ppObject **uintptr) uintptr {
	if IsEqualIID(riid, &IID_IHostMemoryManager) {
		*ppObject = (*uintptr)(unsafe.Pointer(this.MemoryManager))
		return S_OK
	}
	if IsEqualIID(riid, &IID_IHostAssemblyManager) {
		assemblyManagerObj := GetNewCustomAssemblyManager()
		assemblyManager := &assemblyManagerObj
		assemblyManager.TargAssemb = this.TargAssembly
		assemblyManager.AssemblyStore = nil
		*ppObject = (*uintptr)(unsafe.Pointer(assemblyManager))
		return S_OK
	}
	*ppObject = nil
	return uintptr(E_NOTIMPL)
}

// Generate a New Custom HostControl
func GetNewCustomIHostControl() MyHostControl {
	var vtable MyHostControl_Vtbl
	var res MyHostControl

	queryInterfaceAddress := windows.NewCallback(MyQueryInterface)
	addRefAddress := windows.NewCallback(MyAddRef)
	releaseAddress := windows.NewCallback(MyRelease)
	setAppDomainManagerAddress := windows.NewCallback(MySetAppDomainManager)
	getHostManhetAddrees := windows.NewCallback(MyGetHostManager)

	vtable.QueryInterface = queryInterfaceAddress
	vtable.AddRef = addRefAddress
	vtable.Release = releaseAddress
	vtable.SetAppDomainManager = setAppDomainManagerAddress
	vtable.GetHostManager = getHostManhetAddrees

	res.PVtbl = &vtable
	res.Count = 0


	return res
}
