package main

import (
	"fmt"
	clr "pwshexec/go-buena-clr"
	"unsafe"
)

func SystemObjectGetType(appDomain *clr.AppDomain, vtObject clr.Variant) (clr.Variant, error) {
	objectType, err := GetType(appDomain, "System.Runtime", "System.Object")
	if err != nil {
		return clr.Variant{}, fmt.Errorf("GetType(System.Object) failed: %w", err)
	}

	getTypeMethod, err := GetMethod(objectType, flags, "GetType", 0)
	if err != nil {
		return clr.Variant{}, fmt.Errorf("GetMethod(GetType) failed: %w", err)
	}
	defer getTypeMethod.Release()

	emptyArray, err := clr.SafeArrayCreate(0, 0, nil)
	return Invoke(getTypeMethod, vtObject, emptyArray), nil
}

func SystemTypeGetProperty(appDomain *clr.AppDomain, vtTypeObject clr.Variant, propertyName string) (clr.Variant, error) {
	typeType, err := GetType(appDomain, "System.Runtime", "System.Type")
	if err != nil {
		return clr.Variant{}, fmt.Errorf("GetType(System.Type) failed: %w", err)
	}

	getPropertyMethod, err := GetMethod(typeType, flags, "GetProperty", 1)
	if err != nil {
		return clr.Variant{},fmt.Errorf("GetMethod(GetProperty) failed: %w", err)
	}
	defer getPropertyMethod.Release()

	var vtPropName clr.Variant
	if err := clr.InitVariantFromString(propertyName, &vtPropName); err != nil {
		return clr.Variant{},fmt.Errorf("InitVariantFromString failed: %w", err)
	}

	args, err := clr.SafeArrayCreateVector(clr.VT_VARIANT, 0, 1)
	if err != nil {
		return clr.Variant{},fmt.Errorf("SafeArrayCreateVector failed: %w", err)
	}

	if err := clr.SafeArrayPutElement(args, 0, unsafe.Pointer(&vtPropName)); err != nil {
		return clr.Variant{},fmt.Errorf("SafeArrayPutElement failed: %w", err)
	}

	return Invoke(getPropertyMethod, vtTypeObject, args), nil
}


func SystemReflectionPropertyInfoGetValue(appDomain *clr.AppDomain, vtPropertyInfo clr.Variant, vtObject clr.Variant, pIndex *clr.SafeArray) (clr.Variant, error) {
	propertyInfoType, err := GetType(appDomain, "System.Reflection", "System.Reflection.PropertyInfo")
	if err != nil {
		return clr.Variant{}, fmt.Errorf("GetType(PropertyInfo) failed: %w", err)
	}

	numArgs := 1
	if pIndex != nil {
		numArgs = 2
	}

	getValueMethod, err := GetMethod(propertyInfoType, flags, "GetValue", numArgs)
	if err != nil {
		return clr.Variant{},fmt.Errorf("GetMethod(GetValue) failed: %w", err)
	}
	defer getValueMethod.Release()

	args, err := clr.SafeArrayCreateVector(clr.VT_VARIANT, 0, uint64(numArgs))
	if err != nil {
		return clr.Variant{},fmt.Errorf("SafeArrayCreateVector failed: %w", err)
	}
	defer clr.SafeArrayDestroy(args)

	if err := clr.SafeArrayPutElement(args, 0, unsafe.Pointer(&vtObject)); err != nil {
		return clr.Variant{},fmt.Errorf("SafeArrayPutElement[0] failed: %w", err)
	}

	if pIndex != nil {
		vtIndexArray := clr.Variant{
			VT:  clr.VT_VARIANT | clr.VT_ARRAY,
			Val: uintptr(unsafe.Pointer(pIndex)), // assuming parray field maps here
		}
		if err := clr.SafeArrayPutElement(args, 1, unsafe.Pointer(&vtIndexArray)); err != nil {
			return clr.Variant{},fmt.Errorf("SafeArrayPutElement[1] failed: %w", err)
		}
	}

	return Invoke(getValueMethod, vtPropertyInfo, args), nil
}

func SystemReflectionPropertyInfoGetValueSimple(appDomain *clr.AppDomain, vtPropertyInfo, vtObject clr.Variant) (clr.Variant, error) {
	return SystemReflectionPropertyInfoGetValue(appDomain, vtPropertyInfo, vtObject, nil)
}


