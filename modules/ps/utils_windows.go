package main

import (
	"fmt"
	"unsafe"

	"github.com/almounah/superdeye"
	"golang.org/x/sys/windows"
)

const ProcessImageFileName = 27
const ProcessWow64Information = 26

type CLIENT_ID struct {
	pid *uintptr
	tid *uintptr
}
type PCLIENT_ID *CLIENT_ID

type OBJECT_ATTRIBUTES struct {
	Length                   uint64
	RootDirectory            windows.Handle
	ObjectName               *windows.NTUnicodeString
	Attributes               uint64
	SecurityDescriptor       *uintptr
	SecurityQualityOfService *uintptr
}
type POBJECT_ATTRIBUTE *OBJECT_ATTRIBUTES

func NtQuerySystemInformation(sysInfoClass int32, sysInfo unsafe.Pointer, sysInfoLen uint32, retLen *uint32) (ntstatus error) {
	ntstatusint, err := superdeye.SuperdSyscall("NtQuerySystemInformation", uintptr(sysInfoClass), uintptr(sysInfo), uintptr(sysInfoLen), uintptr(unsafe.Pointer(retLen)))
	if err != nil {
		return nil
	}
	return windows.NTStatus(ntstatusint)
}

func NtQueryInformationProcess(proc windows.Handle, procInfoClass int32, procInfo unsafe.Pointer, procInfoLen uint32, retLen *uint32) (ntstatus error) {
	ntstatusint, err := superdeye.SuperdSyscall("NtQueryInformationProcess", uintptr(proc), uintptr(procInfoClass), uintptr(procInfo), uintptr(procInfoLen), uintptr(unsafe.Pointer(retLen)))
	if err != nil {
		return nil
	}
	if ntstatusint == 0 {
		return nil
	}
	return windows.NTStatus(ntstatusint)
}

func NtOpenProcess(desiredAccess uint32, inheritHandle bool, pid uint32) (windows.Handle, error) {
	oa := OBJECT_ATTRIBUTES{}
	cid := CLIENT_ID{(*uintptr)(unsafe.Pointer(uintptr(pid))), nil}

	var out uintptr
	ntstatusint, err := superdeye.SuperdSyscall("NtOpenProcess", uintptr(unsafe.Pointer(&out)), uintptr(desiredAccess), uintptr(unsafe.Pointer(&oa)), uintptr(unsafe.Pointer(&cid)))
	if ntstatusint != 0 {
		return windows.Handle(0), fmt.Errorf("Error NTSTATUS %x", ntstatusint)
	}
	if err != nil {
		return windows.Handle(0), err
	}

	return windows.Handle(out), nil

}

func NtOpenProcessToken(process windows.Handle, access uint32, token *windows.Token) (err error) {
	ntstatusint, err := superdeye.SuperdSyscall("NtOpenProcessToken",uintptr(process), uintptr(access), uintptr(unsafe.Pointer(token)))
	if err != nil {
		return nil
	}
	if ntstatusint == 0 {
		return nil
	}
	return windows.NTStatus(ntstatusint)
}

func NtQueryInformationToken(token windows.Token, infoClass uint32, info *byte, infoLen uint32, returnedLen *uint32) (err error) {
	ntstatusint, err := superdeye.SuperdSyscall("NtQueryInformationToken", uintptr(token), uintptr(infoClass), uintptr(unsafe.Pointer(info)), uintptr(infoLen), uintptr(unsafe.Pointer(returnedLen)), 0)
	if err != nil {
		return nil
	}
	if ntstatusint == 0 {
		return nil
	}
	return windows.NTStatus(ntstatusint)
}
