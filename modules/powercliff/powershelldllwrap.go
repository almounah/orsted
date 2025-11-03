package main

import (
	memorymodule "pwshexec/MemoryModule"
	clr "pwshexec/go-buena-clr"
	"syscall"
	"unsafe"

	_ "embed"

	"golang.org/x/sys/windows"
)

//go:embed powershell.dll
var myDll []byte

var (
	modPowerShell = memorymodule.LoadLibrary(myDll)

	pProcCreatePowerShellConsole                  = memorymodule.GetProcAddress(modPowerShell, "CreatePowerShellConsole")
	pProcExecutePowerShellScript                  = memorymodule.GetProcAddress(modPowerShell, "ExecutePowerShellScript")
	pProcDisablePowerShellEtwProvider             = memorymodule.GetProcAddress(modPowerShell, "DisablePowerShellEtwProvider")
	pProcPatchAllTheThings                        = memorymodule.GetProcAddress(modPowerShell, "PatchAllTheThings")
	pProcCreateInitialRunspaceConfiguration       = memorymodule.GetProcAddress(modPowerShell, "CreateInitialRunspaceConfiguration")
	pProcStartConsoleShell                        = memorymodule.GetProcAddress(modPowerShell, "StartConsoleShell")
	pProcPowerShellCreate                         = memorymodule.GetProcAddress(modPowerShell, "PowerShellCreate")
	pProcPowerShellDispose                        = memorymodule.GetProcAddress(modPowerShell, "PowerShellDispose")
	pProcPowerShellAddScript                      = memorymodule.GetProcAddress(modPowerShell, "PowerShellAddScript")
	pProcPowerShellAddCommand                     = memorymodule.GetProcAddress(modPowerShell, "PowerShellAddCommand")
	pProcPowerShellInvoke                         = memorymodule.GetProcAddress(modPowerShell, "PowerShellInvoke")
	pProcPowerShellClear                          = memorymodule.GetProcAddress(modPowerShell, "PowerShellClear")
	pProcGetJustInTimeMethodAddress               = memorymodule.GetProcAddress(modPowerShell, "GetJustInTimeMethodAddressEx")
	pProcPowerShellClearErrors                    = memorymodule.GetProcAddress(modPowerShell, "PowerShellClearErrors")
	pProcPowerShellGetStream                      = memorymodule.GetProcAddress(modPowerShell, "PowerShellGetStream")
	pProcPowerShellHadErrors                      = memorymodule.GetProcAddress(modPowerShell, "PowerShellHadErrors")
	pProcPrintPowerShellInvokeResult              = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellInvokeResult")
	pProcPrintPowerShellInvokeInformation         = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellInvokeInformation")
	pProcPrintPowerShellInvokeErrors              = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellInvokeErrors")
	pProcPrintInformationRecord                   = memorymodule.GetProcAddress(modPowerShell, "PrintInformationRecord")
	pProcPrintErrorRecord                         = memorymodule.GetProcAddress(modPowerShell, "PrintErrorRecord")
	pProcPrintPowerShellInformationStream         = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellInformationStream")
	pProcPrintPowerShellErrorStream               = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellErrorStream")
	pProcPrintPowerShellInvocationStateInfoReason = memorymodule.GetProcAddress(modPowerShell, "PrintPowerShellInvocationStateInfoReason")
	pProcSetConsoleTextColor                      = memorymodule.GetProcAddress(modPowerShell, "SetConsoleTextColor")
)

// func CreatePowerShellConsole()
// func ExecutePowerShellScript(pwszScript *uint16)
//
// func DisablePowerShellEtwProvider(pAppDomain *clr.AppDomain) bool
// func PatchAllTheThings(pAppDomain *clr.AppDomain)
//
// func CreateInitialRunspaceConfiguration(pAppDomain *clr.AppDomain, pvtRunspaceConfiguration *clr.Variant) bool
// func StartConsoleShell(pAppDomain *clr.AppDomain, vtRunspaceConfiguration clr.Variant, pwszBanner *uint16, pwszHelp *uint16, ppwszArguments **uint16, dwArgumentCount uint32) bool
func PowerShellCreate(pAppDomain *clr.AppDomain) (clr.Variant, error) {
	var result clr.Variant
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellCreate,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&result)),
	)
	if ret == 0 {
		return clr.Variant{}, syscall.Errno(ret)
	}
	return result, nil
}

func PowerShellDispose(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (bool, error) {
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellDispose,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
	)
	if ret == 0 {
		return false, syscall.Errno(ret)
	}
	return true, nil
}

func PowerShellAddScript(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant, script string) (bool, error) {
	ptr, err := syscall.UTF16PtrFromString(script)
	if err != nil {
		return false, err
	}
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellAddScript,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(ptr)),
	)
	if ret == 0 {
		return false, syscall.Errno(ret)
	}
	return true, nil
}

func PowerShellAddCommand(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant, command string) (bool, error) {
	ptr, err := syscall.UTF16PtrFromString(command)
	if err != nil {
		return false, err
	}
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellAddCommand,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(ptr)),
	)
	if ret == 0 {
		return false, syscall.Errno(ret)
	}
	return true, nil
}

func PowerShellInvoke(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (clr.Variant, error) {
	var invokeResult clr.Variant
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellInvoke,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&invokeResult)),
	)
	if ret == 0 {
		return clr.Variant{}, syscall.Errno(ret)
	}
	return invokeResult, nil
}

func PowerShellClear(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (clr.Variant, error) {
	var invokeResult clr.Variant
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellClear,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&invokeResult)),
	)
	if ret == 0 {
		return clr.Variant{}, syscall.Errno(ret)
	}
	return invokeResult, nil
}

func PowerShellClearErrors(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (clr.Variant, error) {
	var invokeResult clr.Variant
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellClearErrors,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&invokeResult)),
	)
	if ret == 0 {
		return clr.Variant{}, syscall.Errno(ret)
	}
	return invokeResult, nil
}

func PowerShellGetStream(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant, streamName string) (clr.Variant, error) {
	var stream clr.Variant
	ptr, err := syscall.UTF16PtrFromString(streamName)
	if err != nil {
		return clr.Variant{}, err
	}
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellGetStream,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&stream)),
	)
	if ret == 0 {
		return clr.Variant{}, syscall.Errno(ret)
	}
	return stream, nil
}

func PowerShellHadErrors(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (bool, error) {
	var hadErrorsInt int32
	ret, _, _ := syscall.SyscallN(
		pProcPowerShellHadErrors,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&hadErrorsInt)),
	)
	if ret == 0 {
		return false, syscall.Errno(ret)
	}
	return hadErrorsInt != 0, nil
}

func PrintPowerShellInvokeResult(pAppDomain *clr.AppDomain, vtInvokeResult clr.Variant) (string, error) {
	var output *uint16
	syscall.SyscallN(
		pProcPrintPowerShellInvokeResult,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtInvokeResult)),
		uintptr(unsafe.Pointer(&output)),
	)
	// If the native function doesn't return a status, we can't detect errors here
	defer windows.CoTaskMemFree(unsafe.Pointer(output))
	return windows.UTF16PtrToString(output), nil
}

func PrintPowerShellInvokeInformation(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (string, error) {
	var output *uint16
	syscall.SyscallN(
		pProcPrintPowerShellInvokeInformation,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&output)),
	)
	// If the native function doesn't return a status, we can't detect errors here
	return windows.UTF16PtrToString(output), nil
}

func PrintPowerShellInvokeErrors(pAppDomain *clr.AppDomain, vtPowerShellInstance clr.Variant) (string, error) {
	var output *uint16
	syscall.SyscallN(
		pProcPrintPowerShellInvokeErrors,
		uintptr(unsafe.Pointer(pAppDomain)),
		uintptr(unsafe.Pointer(&vtPowerShellInstance)),
		uintptr(unsafe.Pointer(&output)),
	)
	// If the native function doesn't return a status, we can't detect errors here
	return windows.UTF16PtrToString(output), nil
}

func GetFunctionAddressJIT(appDomain *clr.AppDomain, assemblyName string, className string, methodName string, nbarg uint32) (address uintptr, err error) {
	Println("Will Call GetFunctionAddressJIT")
	pAssName, err := syscall.UTF16PtrFromString(assemblyName)
	if err != nil {
		return 0, err
	}
	pclassName, err := syscall.UTF16PtrFromString(className)
	if err != nil {
		return 0, err
	}
	pmethodName, err := syscall.UTF16PtrFromString(methodName)
	if err != nil {
		return 0, err
	}
	var res uintptr = 0
	var nb = nbarg
	Println("Will do syscall on ", pProcGetJustInTimeMethodAddress)
	ret, _, _ := syscall.SyscallN(
		pProcGetJustInTimeMethodAddress,
		uintptr(unsafe.Pointer(appDomain)),
		uintptr(unsafe.Pointer(pAssName)),
		uintptr(unsafe.Pointer(pclassName)),
		uintptr(unsafe.Pointer(pmethodName)),
		uintptr(unsafe.Pointer(&nb)),
		uintptr(unsafe.Pointer(&res)),
	)
	Println("Done syscall ->", ret)
	if ret != 0 && ret != 1 {
		return res, syscall.Errno(ret)
	}
	return res, nil
}

//func PrintPowerShellInvocationStateInfoReason(pAppDomain *clr.AppDomain, vtReason clr.Variant)

//func PrintInformationRecord(pAppDomain *clr.AppDomain, vtInformationRecord clr.Variant)
//func PrintErrorRecord(pAppDomain *clr.AppDomain, vtErrorRecord clr.Variant)
//func PrintPowerShellInformationStream(pAppDomain *clr.AppDomain, vtInformationStream clr.Variant)
//func PrintPowerShellErrorStream(pAppDomain *clr.AppDomain, vtErrorStream clr.Variant)
//func SetConsoleTextColor(wColor uint16, pwOldColor *uint16)
