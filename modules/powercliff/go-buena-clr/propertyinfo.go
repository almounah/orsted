package clr

type PropertyInfo struct {
	Vtlb *PropertyInfoVtbl
}

type PropertyInfoVtbl struct {
	QueryInterface        uintptr
	AddRef                uintptr
	Release               uintptr
	GetTypeInfoCount      uintptr
	GetTypeInfo           uintptr
	GetIDsOfNames         uintptr
	Invoke                uintptr
	ToString              uintptr
	Equals                uintptr
	GetHashCode           uintptr
	GetType               uintptr
	get_MemberType        uintptr
	get_name              uintptr
	get_DeclaringType     uintptr
	get_ReflectedType     uintptr
	GetCustomAttributes   uintptr
	GetCustomAttributes_2 uintptr
	IsDefined             uintptr
	get_PropertyType      uintptr
	GetValue              uintptr
	GetValue_2            uintptr
	SetValue              uintptr
	SetValue_2            uintptr
	GetAccessors          uintptr
	GetGetMethod          uintptr
	GetSetMethod          uintptr
	GetIndexParameters    uintptr
	get_Attributes        uintptr
	get_CanRead           uintptr
	get_CanWrite          uintptr
	GetAccessors_2        uintptr
	GetGetMethod_2        uintptr
	GetSetMethod_2        uintptr
	get_IsSpecialName     uintptr
}
