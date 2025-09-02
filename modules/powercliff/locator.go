package main

import (
	"fmt"
	"os"
	"path/filepath"
	clr "pwshexec/go-buena-clr"

	"golang.org/x/sys/windows"

	"syscall"
	"unsafe"
)

const (
	BindingFlags_Instance     = 0x00000004
	BindingFlags_Static       = 0x00000008
	BindingFlags_Public       = 0x00000010
	BindingFlags_NonPublic    = 0x00000020
	BindingFlags_DeclaredOnly = 0x00000002
)

var flags = BindingFlags_Instance |
	BindingFlags_Static |
	BindingFlags_Public |
	BindingFlags_NonPublic |
	BindingFlags_DeclaredOnly

func FindAssemblyPath(assemblyName string) (string, error) {
	const gacPath = `C:\Windows\Microsoft.NET\assembly\GAC_MSIL`

	assemblyFolder := filepath.Join(gacPath, assemblyName)

	entries, err := os.ReadDir(assemblyFolder)
	if err != nil {
		return "", fmt.Errorf("failed to read GAC folder: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fullPath := filepath.Join(assemblyFolder, entry.Name(), assemblyName+".dll")
			if _, err := os.Stat(fullPath); err == nil {
				return fullPath, nil
			}
		}
	}

	return "", fmt.Errorf("assembly %q not found in GAC", assemblyName)
}

func LoadAssembly(appDomain *clr.AppDomain, assemblyName string) (*clr.Assembly, error) {
	if assemblyName == "mscorlib" {
		path := "C:\\Windows\\Microsoft.NET\\Framework64\\v4.0.30319\\mscorlib.dll"
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("Error in ReadFile, %s", err)
		}
		sa, err := clr.CreateSafeArray(b)
		if err != nil {
			return nil, fmt.Errorf("Error in CreateSafeArrauy, %s", err)
		}

		assemb, err := appDomain.Load_3(sa)

		if err != nil {
		}
		return assemb, nil
	}
	path, err := FindAssemblyPath(assemblyName)
	if err != nil {
		return nil, fmt.Errorf("Error in FindAssembylPath, %s", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error in ReadFile, %s", err)
	}
	sa, err := clr.CreateSafeArray(b)
	if err != nil {
		return nil, fmt.Errorf("Error in CreateSafeArrauy, %s", err)
	}

	assemb, err := appDomain.Load_3(sa)

	if err != nil {
	}

	return assemb, nil

}

func GetType(appDomain *clr.AppDomain, assemblyName string, typeFullName string) (*clr.SystemType, error) {
	assembly, err := LoadAssembly(appDomain, assemblyName)
	if err != nil {
		return nil, fmt.Errorf("Error in LoadAssembly, %s", err)
	}

	var systemType *clr.SystemType
	s1, err := clr.SysAllocString(typeFullName)
	if err != nil {
		return nil, err
	}
	hr, _, _ := syscall.SyscallN(assembly.Vtbl.GetType_2, uintptr(unsafe.Pointer(assembly)), uintptr(s1), uintptr(unsafe.Pointer(&systemType)))
	if hr != 0 {
	}
	return systemType, nil
}

func FindMethodInArray(pMethods *clr.SafeArray, methName string, nbrArgs int) (*clr.MethodInfo, error) {
	ppMethods, err := clr.SafeArrayAccessData(pMethods)
	if err != nil {
		return nil, err
	}
	lowerBound, err := clr.SafeArrayGetLBound(pMethods, 1)
	if err != nil {
		return nil, err
	}
	UpperBound, err := clr.SafeArrayGetUBound(pMethods, 1)
	if err != nil {
		return nil, err
	}


	methodArray := unsafe.Slice((**clr.MethodInfo)(unsafe.Pointer(ppMethods)), UpperBound-lowerBound+1)
	for _, meth := range methodArray {
		var bName uintptr
		syscall.SyscallN(meth.Vtbl.Get_name, uintptr(unsafe.Pointer(meth)), uintptr(unsafe.Pointer(&bName)))
		goName := windows.UTF16PtrToString((*uint16)(unsafe.Pointer(bName)))
		if goName == methName {
			return meth, nil
		}

	}

	return nil, fmt.Errorf("Cannot find method name")
}

func GetMethod(systemType *clr.SystemType, flags int, methName string, nbArg int) (*clr.MethodInfo, error) {
	var pMethods *clr.SafeArray
	hr, _, _ := syscall.SyscallN(systemType.Vtbl.GetMethods, uintptr(unsafe.Pointer(systemType)), uintptr(flags), uintptr(unsafe.Pointer(&pMethods)))
	if hr != 0 {
	}
	return FindMethodInArray(pMethods, methName, nbArg)
}

func GetProperty(pType *clr.SystemType, bindingFlags uint32, vtObject clr.Variant, propertyName string) (*clr.PropertyInfo, error) {
	bString, err := clr.SysAllocString(propertyName)
	if err != nil {
		return nil, err
	}
	var propInfo *clr.PropertyInfo
	hr, _, _ := syscall.SyscallN(pType.Vtbl.GetProperty, uintptr(unsafe.Pointer(pType)), uintptr(unsafe.Pointer(bString)), uintptr(bindingFlags), uintptr(unsafe.Pointer(&propInfo)))
	if hr != 0 {
	}
	return propInfo, nil
}

func GetPropertyValue(pType *clr.SystemType, bindingFlags uint32, vtObject clr.Variant, propertyName string) (clr.Variant, error) {

	propInfo, err := GetProperty(pType, bindingFlags, vtObject, propertyName)
	if err != nil {
		return clr.Variant{}, err
	}

	var vtResult clr.Variant = clr.Variant{}
	hr, _, _ := syscall.SyscallN(propInfo.Vtlb.GetValue, uintptr(unsafe.Pointer(propInfo)), uintptr(unsafe.Pointer(&vtObject)), 0, uintptr(unsafe.Pointer(&vtResult)))
	if hr != 0 {
	}
	return vtResult, nil
}

func Invoke(methodInfo *clr.MethodInfo, vtObject clr.Variant, param *clr.SafeArray) clr.Variant {
	var res clr.Variant
	hr, _, _ := syscall.SyscallN(methodInfo.Vtbl.Invoke_3, uintptr(unsafe.Pointer(methodInfo)), uintptr(unsafe.Pointer(&vtObject)), uintptr(unsafe.Pointer(param)), uintptr(unsafe.Pointer(&res)))
	if hr != 0 {
	}
	return res

}
func PrepareMethod(appDomain *clr.AppDomain, pvtMethodHandle *clr.Variant) {
	pRuntimeHelpersType, _ := GetType(appDomain, "System.Runtime", "System.Runtime.CompilerServices.RuntimeHelpers")
	_, _ = GetMethod(pRuntimeHelpersType, flags, "PrepareMethod", 0)
	prepMethodArgument, _ := clr.SafeArrayCreateVector(clr.VT_VARIANT, 0, 1)


	lpArgyument := 0
	clr.SafeArrayPutElement(prepMethodArgument, int32(lpArgyument), unsafe.Pointer(pvtMethodHandle))
}

func GetFunctionPointer(appDomain *clr.AppDomain, pvtMethiodHandle *clr.Variant) (uintptr, error) {
	pRuntimeMethodHandleType, _ := GetType(appDomain, "System.Runtime", "System.RuntimeMethodHandle")
	pGetFunctionPointerInfo, _ := GetMethod(pRuntimeMethodHandleType, flags, "GetFunctionPointer", 0)
	getFuncPtrArgs, err := clr.SafeArrayCreate(0, 0, nil)
	if err != nil {
	}

	res := Invoke(pGetFunctionPointerInfo, *pvtMethiodHandle, getFuncPtrArgs)

	addr := res.Val
	return addr, nil
}

func GetFunctionAddress(appDomain *clr.AppDomain, assemblyName string, className string, methodName string, nbarg uint32) (address uintptr, err error) {
	sType, err := GetType(appDomain, assemblyName, className)
	if err != nil {
		return 0, fmt.Errorf("Error in GetType, %s", err)
	}
	pTargetMethodInfo, err := GetMethod(sType, flags, methodName, int(nbarg))
	if err != nil {
		return 0, fmt.Errorf("Error in GetMethod for targetmnethodinfo, %s", err)
	}

	// Getting handle over this method
	pMethoInfoType, err := GetType(appDomain, "System.Reflection", "System.Reflection.MethodInfo")
	if err != nil {
		return 0, fmt.Errorf("Error in GetType for method info type, %s", err)
	}

	vtMethodHandler := clr.Variant{}
	vtMethodHandler.VT = 13
	vtMethodHandler.Val = uintptr(unsafe.Pointer(pTargetMethodInfo))

	vtMethodHandleVal, err := GetPropertyValue(pMethoInfoType, 4|16, vtMethodHandler, "MethodHandle")
	if err != nil {
		return 0, err
	}

	// Get Effective Address
	PrepareMethod(appDomain, &vtMethodHandleVal)
	addr, err := GetFunctionPointer(appDomain, &vtMethodHandleVal)
	return addr, nil
}
