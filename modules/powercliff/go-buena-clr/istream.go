package clr

import (
	"syscall"
	"unsafe"
)

type IStream struct {
	Vtbl *IStreamVtbl
}

type IStreamVtbl struct {
	QueryInterface uintptr
	AddRef         uintptr
	Release        uintptr
	Read           uintptr
	Write          uintptr
	Seek           uintptr
	SetSize        uintptr
	CopyTo         uintptr
	Commit         uintptr
	Revert         uintptr
	LockRegion     uintptr
	UnlockRegion   uintptr
	Stat           uintptr
	Clone          uintptr
}


// Read reads from the stream into a buffer
func (obj *IStream) Read(buffer []byte) (uint32, error) {
	var bytesRead uint32
	ret, _, _ := syscall.SyscallN(obj.Vtbl.Read,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&bytesRead)))

	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return bytesRead, nil
}

// Write writes to the stream
func (obj *IStream) Write(buffer []byte) (uint32, error) {
	var bytesWritten uint32
	ret, _, _ := syscall.SyscallN(obj.Vtbl.Write,
		uintptr(unsafe.Pointer(obj)),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
		uintptr(unsafe.Pointer(&bytesWritten)))

	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return bytesWritten, nil
}

// Seek moves the stream pointer
func (obj *IStream) Seek(offset int64, origin uint32) (uint64, error) {
	var newPos uint64
	ret, _, _ := syscall.SyscallN(obj.Vtbl.Seek,
		uintptr(unsafe.Pointer(obj)),
		uintptr(offset),
		uintptr(origin),
		uintptr(unsafe.Pointer(&newPos)))

	if ret != 0 {
		return 0, syscall.Errno(ret)
	}
	return newPos, nil
}

// Release decrements the reference count
func (obj *IStream) Release() uint32 {
	ret, _, _ := syscall.SyscallN(obj.Vtbl.Release, uintptr(unsafe.Pointer(obj)))
	return uint32(ret)
}
