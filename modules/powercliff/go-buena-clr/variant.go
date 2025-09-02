//go:build windows
// +build windows

package clr

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	// VT_EMPTY No value was specified. If an optional argument to an Automation method is left blank, do not
	// pass a VARIANT of type VT_EMPTY. Instead, pass a VARIANT of type VT_ERROR with a value of DISP_E_PARAMNOTFOUND.
	VT_EMPTY uint16 = 0x0000
	// VT_NULL A propagating null value was specified. (This should not be confused with the null pointer.)
	// The null value is used for tri-state logic, as with SQL.
	VT_NULL uint16 = 0x0001
	// VT_UI1 is a Variant Type of Unsigned Integer of 1-byte
	VT_UI1 uint16 = 0x0011
	// VT_UT4 is a Varriant Type of Unsigned Integer of 4-byte
	VT_UI4 uint16 = 0x0013
	// VT_BSTR is a Variant Type of BSTR, an OLE automation type for transfering length-prefixed strings
	// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-oaut/9c5a5ce4-ff5b-45ce-b915-ada381b34ac1
	VT_BSTR uint16 = 0x0008
	// VT_VARIANT is a Variant Type of VARIANT, a container for a union that can hold many types of data
	VT_VARIANT uint16 = 0x000c
	// VT_ARRAY is a Variant Type of a SAFEARRAY
	// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-oaut/2e87a537-9305-41c6-a88b-b79809b3703a
	VT_ARRAY uint16 = 0x2000
)

// from https://github.com/go-ole/go-ole/blob/master/variant_amd64.go
// https://docs.microsoft.com/en-us/windows/win32/winauto/variant-structure
// https://docs.microsoft.com/en-us/windows/win32/api/oaidl/ns-oaidl-variant
// https://docs.microsoft.com/en-us/previous-versions/windows/embedded/ms891678(v=msdn.10)?redirectedfrom=MSDN
// VARIANT Type Constants https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-oaut/3fe7db9f-5803-4dc4-9d14-5425d3f5461f
type Variant struct {
	VT         uint16 // VARTYPE
	wReserved1 uint16
	wReserved2 uint16
	wReserved3 uint16
	Val        uintptr
	_          [8]byte
}

func InitVariantFromString(s string, pv *Variant) error {
	return InitVariantFromStringArray([]string{s}, pv)
}

func InitVariantFromBoolean(val bool, pv *Variant) error {
	return InitVariantFromBooleanArray([]bool{val}, pv)
}

func InitVariantFromInt32(val int32, pv *Variant) error {
	return InitVariantFromInt32Array([]int32{val}, pv)
}

func InitVariantFromStringArray(arr []string, pv *Variant) error {
	propsys := syscall.NewLazyDLL("propsys.dll")
	procInitVariantFromStringVector := propsys.NewProc("InitVariantFromStringArray")

	count := len(arr)
	ptrs := make([]*uint16, count)

	for i, s := range arr {
		utf16Str, err := syscall.UTF16PtrFromString(s)
		if err != nil {
			return fmt.Errorf("UTF16PtrFromString failed at index %d: %w", i, err)
		}
		ptrs[i] = utf16Str
	}

	hr, _, _ := procInitVariantFromStringVector.Call(
		uintptr(unsafe.Pointer(&ptrs[0])),
		uintptr(count),
		uintptr(unsafe.Pointer(pv)),
	)

	if hr != S_OK {
		return fmt.Errorf("InitVariantFromStringVector failed: HRESULT=0x%x", hr)
	}
	return nil
}


func InitVariantFromBooleanArray(arr []bool, pv *Variant) error {
	propsys := syscall.NewLazyDLL("propsys.dll")
	procInitPropVariantFromBooleanVector := propsys.NewProc("InitVariantFromBooleanArray")

	n := len(arr)
	if n == 0 {
		return fmt.Errorf("array must not be empty")
	}

	// Convert Go []bool → []int32 (Windows BOOL)
	boolArray := make([]int32, n)
	for i, v := range arr {
		if v {
			boolArray[i] = -1 // VARIANT_TRUE = -1
		} else {
			boolArray[i] = 0 // VARIANT_FALSE
		}
	}

	hr, _, _ := procInitPropVariantFromBooleanVector.Call(
		uintptr(unsafe.Pointer(&boolArray[0])), // *BOOL
		uintptr(n),                             // count
		uintptr(unsafe.Pointer(pv)),            // PROPVARIANT out
	)

	if hr != S_OK {
		return fmt.Errorf("InitPropVariantFromBooleanVector failed: HRESULT=0x%x", hr)
	}
	return nil
}

func InitVariantFromInt32Array(values []int32, pv *Variant) error {
	propsys := syscall.NewLazyDLL("propsys.dll")
	procInitVariantFromInt32Array := propsys.NewProc("InitVariantFromBooleanArray")
	if len(values) == 0 {
		return fmt.Errorf("empty array")
	}

	ret, _, err := procInitVariantFromInt32Array.Call(
		uintptr(unsafe.Pointer(&values[0])),
		uintptr(len(values)),
		uintptr(unsafe.Pointer(pv)),
	)
	if ret != 0 { // S_OK = 0
		return fmt.Errorf("InitVariantFromInt32Array failed: HRESULT=0x%x, %v", ret, err)
	}
	return nil
}


func VariantClear(psa *Variant) error {
	debugPrint("Entering into safearray.SafeArrayDestroy()...")

	modOleAuto := syscall.MustLoadDLL("OleAut32.dll")
	safeArrayDestroy := modOleAuto.MustFindProc("VariantClear")

	hr, _, err := safeArrayDestroy.Call(
		uintptr(unsafe.Pointer(psa)),
		0,
		0,
	)

	if err != syscall.Errno(0) {
		return fmt.Errorf("the oleaut32!VariantClear function call returned an error:\n%s", err)
	}
	if hr != S_OK {
		return fmt.Errorf("the oleaut32!VariantClear function returned a non-zero HRESULT: 0x%x", hr)
	}
	return nil
}
