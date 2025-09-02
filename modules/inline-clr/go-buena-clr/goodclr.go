package clr

import (
	"fmt"
	"syscall"
	"unsafe"


	"golang.org/x/sys/windows"
)

func LoadGoodClr(runtime string, NetBytes []byte) (*ICORRuntimeHost, []uint16, error) {
	var myCustomHost *ICLRRuntimeHost

	// Get the Metahost
	pMetahost, err := CLRCreateInstance(CLSID_CLRMetaHost, IID_ICLRMetaHost)
	if err != nil {
		return nil, nil, err
	}

	// Get Runtime Info from the Metahost
	pRuntimeInfo, err := GetRuntimeInfo(pMetahost, runtime)
	if err != nil {
		return nil, nil, err
	}

	// Get the ICLRRuntimeHost and store it in MyCustomHost
	err = pRuntimeInfo.GetInterface(CLSID_CLRRuntimeHost, IID_ICLRRuntimeHost, unsafe.Pointer(&myCustomHost))
	if err != nil {
		return nil, nil, err
	}


	var pIdentityManagerProc uintptr
	var pIdentityManager *ICLRAssemblyIdentityManager
	res, _ := windows.BytePtrFromString("GetCLRIdentityManager")
	syscall.SyscallN(pRuntimeInfo.vtbl.GetProcAddress, uintptr(unsafe.Pointer(pRuntimeInfo)), uintptr(unsafe.Pointer(res)), uintptr(unsafe.Pointer(&pIdentityManagerProc)))

	syscall.SyscallN(pIdentityManagerProc, uintptr(unsafe.Pointer(&IID_ICLRAssemblyIdentityManager)), uintptr(unsafe.Pointer(&pIdentityManager)))

	// Load Assembly
	assemblyStream := SHCreateMemStream(&NetBytes[0], uint32(len(NetBytes)))
	identityBuffer := make([]uint16, 4096)
	identityBufferSize := uint32(4096)
	syscall.SyscallN(pIdentityManager.VTable.GetBindingIdentityFromStream, uintptr(unsafe.Pointer(pIdentityManager)), uintptr(unsafe.Pointer(assemblyStream)), uintptr(0), uintptr(unsafe.Pointer(&identityBuffer[0])), uintptr(unsafe.Pointer(&identityBufferSize)))

	fmt.Println(fmt.Sprintf("[+] GetBindingIdentityFromStream: %s", syscall.UTF16ToString(identityBuffer)))

    // Create a Target Assembly and bind it to our Host
    var tgtAssembly *TargetAssembly = &TargetAssembly{}
    pG, _ := syscall.UTF16PtrFromString(syscall.UTF16ToString(identityBuffer))
    tgtAssembly.AssemblyInfo = pG
    tgtAssembly.AssemblyBytes = &NetBytes[0]
    tgtAssembly.AssemblySize = uint32(len(NetBytes))

    fmt.Println("[?] Printing TargetAssembly otherwise it dies because of Go GC", tgtAssembly)

    // Create Memory Manager
    // memManger := GetMemoryManager()
    

	// Initialise our custom IHOSTControl
	var customHostControl MyHostControl
	customHostControl = GetNewCustomIHostControl()
    customHostControl.TargAssembly = tgtAssembly
    //customHostControl.MemoryManager = memManger

	// Check Identity Manager That we will Use Later
	syscall.SyscallN(myCustomHost.vtbl.SetHostControl, uintptr(unsafe.Pointer(myCustomHost)), uintptr(unsafe.Pointer(&customHostControl)))
	fmt.Println("[+] Set Custom Host Successfully")
    
	myCustomHost.Start()
	fmt.Println("[+] Started CLR Succesfully")

	var runtimeHost *ICORRuntimeHost
	err = pRuntimeInfo.GetInterface(CLSID_CorRuntimeHost, IID_ICorRuntimeHost, unsafe.Pointer(&runtimeHost))
	return runtimeHost, identityBuffer, nil
}

func Load2Assembly(runtimeHost *ICORRuntimeHost, identityString []uint16) (*Assembly){
	appDomain, err := GetAppDomain(runtimeHost)
	if err != nil {
		return nil
	}

    // Patchin System Exit
    //PatchSysExit(appDomain)

    var assembly *Assembly 
    s := windows.UTF16ToString(identityString)
    assembly, _ = appDomain.Load_2(s)
	return assembly
}

func PatchSysExit(appDomain *AppDomain) {
    mscorlib, err := appDomain.Load_2("mscorlib, Version=4.0.0.0")
    if err != nil {
        fmt.Println("Error Returning ", err)
    }

    // Get The Exit Class
    var exitClass *SystemType
    s1, _ := SysAllocString("System.Environment")
    hr, _, e := syscall.SyscallN(mscorlib.vtbl.GetType_2, uintptr(unsafe.Pointer(mscorlib)), uintptr(s1), uintptr(unsafe.Pointer(&exitClass)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetType ---> %X, %s", hr, e))

    // Get The Exit Method
    var exitInfo *MethodInfo
    s2, _ := SysAllocString("Exit")
    exitFlags := 16 | 8
    hr, _, e = syscall.SyscallN(exitClass.vtbl.GetMethod_2, uintptr(unsafe.Pointer(exitClass)), uintptr(s2), uintptr(exitFlags), uintptr(unsafe.Pointer(&exitInfo)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall ExitClass.GetMethod_2 ---> %X, %s", hr, e))

    // Getting methodInfoClass
    var methodInfoClass *SystemType
    s3, _ := SysAllocString("System.Reflection.MethodInfo")
    hr, _, e = syscall.SyscallN(mscorlib.vtbl.GetType_2, uintptr(unsafe.Pointer(mscorlib)), uintptr(s3), uintptr(unsafe.Pointer(&methodInfoClass)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetType methodInfoClass ---> %X, %s", hr, e))

    // Get Property Info
    var methodHandleProp *PropertyInfo
    s4, _ := SysAllocString("MethodHandle")
    methodHandleFlags := 4 | 16
    hr, _, e = syscall.SyscallN(methodInfoClass.vtbl.GetProperty, uintptr(unsafe.Pointer(methodInfoClass)), uintptr(s4), uintptr(methodHandleFlags), uintptr(unsafe.Pointer(&methodHandleProp)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetProperty ---> %X, %s", hr, e))

    // Some Stuff With Variant
    var methodHandlePtr Variant
    methodHandlePtr.VT = 13
    methodHandlePtr.Val = uintptr(unsafe.Pointer(exitInfo))

    methodHandleArgs, err := SafeArrayCreate(0, 0, nil)
    if err != nil {
        fmt.Println("Suspicious correct")
    }
    methodHandleVal := Variant{}
    hr, _, e = syscall.SyscallN(methodHandleProp.Vtlb.GetValue, uintptr(unsafe.Pointer(methodHandleProp)), uintptr(unsafe.Pointer(&methodHandlePtr)), uintptr(unsafe.Pointer(methodHandleArgs)), uintptr(unsafe.Pointer(&methodHandleVal)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetValue ---> %X, %s", hr, e))

    // Get GetFunctionPointer function
    var rtMethodHandleType *SystemType
    s5, _ := SysAllocString("System.RuntimeMethodHandle")
    hr, _, e = syscall.SyscallN(mscorlib.vtbl.GetType_2, uintptr(unsafe.Pointer(mscorlib)), uintptr(s5), uintptr(unsafe.Pointer(&rtMethodHandleType)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetType ---> %X, %s", hr, e))

    var getFuncPtrMethodInfo *MethodInfo
    s2, _ = SysAllocString("GetFunctionPointer")
    getFuncPtrFlags := 4 | 16
    hr, _, e = syscall.SyscallN(rtMethodHandleType.vtbl.GetMethod_2, uintptr(unsafe.Pointer(rtMethodHandleType)), uintptr(s2), uintptr(getFuncPtrFlags), uintptr(unsafe.Pointer(&getFuncPtrMethodInfo)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall ExitClass.GetMethod_2 ---> %X, %s", hr, e))

    getFuncPtrArgs, err := SafeArrayCreate(0, 0, nil)
    if err != nil {
        fmt.Println("Suspicious correct")
    }
    exitPtr := Variant{}
    hr, _, e = syscall.SyscallN(getFuncPtrMethodInfo.vtbl.Invoke_3, uintptr(unsafe.Pointer(getFuncPtrMethodInfo)), uintptr(unsafe.Pointer(&methodHandleVal)), uintptr(unsafe.Pointer(getFuncPtrArgs)), uintptr(unsafe.Pointer(&exitPtr)))
    fmt.Println(fmt.Sprintf("[+] Done Syscall GetValue ---> %X, %s", hr, e))

    addr := exitPtr.Val
    fmt.Println(exitPtr)
    fmt.Println(fmt.Sprintf("Address is %X", exitPtr.Val))
    var oldpro uint32
    err = windows.VirtualProtect(addr, 1, windows.PAGE_READWRITE, &oldpro)
    if err != nil {
        fmt.Println("Error In VirtualProtect")
        fmt.Println(err.Error())
    }

    patch := []byte{0xC3}
    Memcpy(unsafe.Pointer(addr), unsafe.Pointer(&patch[0]), 1)

    var oldpro2 uint32
    err = windows.VirtualProtect(addr, 1, oldpro, &oldpro2)
    if err != nil {
        fmt.Println("Error In VirtualProtect")
        fmt.Println(err.Error())
    }

}
