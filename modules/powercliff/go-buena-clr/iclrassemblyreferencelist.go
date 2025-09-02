package clr

type ICLRAssemblyReferenceList struct {
	VTable *ICLRAssemblyReferenceListVTable
}

type ICLRAssemblyReferenceListVTable struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	AddReference   uintptr
	GetCount       uintptr
	GetReferenceAt uintptr
}
