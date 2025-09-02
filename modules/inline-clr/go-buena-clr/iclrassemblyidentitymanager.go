package clr

type ICLRAssemblyIdentityManager struct {
	VTable *ICLRAssemblyIdentityManagerVTable
}

type ICLRAssemblyIdentityManagerVTable struct {
	QueryInterface                    uintptr // Correct Order
	AddRef                            uintptr // Correct Order
	Release                           uintptr // Correct Order
	GetCLRAssemblyReferenceList       uintptr
	GetBindingIdentityFromFile        uintptr // Correct Order
	GetBindingIdentityFromStream      uintptr // Correct Order
	GetProbingAssembliesFromReference uintptr
	GetReferencedAssembliesFromFile   uintptr
	GetReferencedAssembliesFromStream uintptr
	IsStronglyNamed                   uintptr
}
