package clr

type TargetAssembly struct {
	AssemblyInfo    *uint16
	AssemblyBytes   *byte
	AssemblySize    uint32
	IAssemblyStream *IStream
}
