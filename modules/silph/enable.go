package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	advapi32                  = windows.NewLazyDLL("advapi32.dll")
	kernel32                  = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcessToken      = advapi32.NewProc("OpenProcessToken")
	procLookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueA")
	procAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
	procRegOpenKeyExA         = advapi32.NewProc("RegOpenKeyExA")
	procRegSaveKeyA           = advapi32.NewProc("RegSaveKeyA")
	procRegCloseKey           = advapi32.NewProc("RegCloseKey")
)

const (
	KEY_ALL_ACCESS            = 0xF003F
	TOKEN_ADJUST_PRIVILEGES   = 0x20
	TOKEN_QUERY               = 0x8
	SE_PRIVILEGE_ENABLED      = 0x2
	REG_OPTION_BACKUP_RESTORE = 0x00000004
)

func enablePriv(priv string) error {
	var hToken syscall.Handle
	currProc := windows.CurrentProcess()
	ret, _, err := procOpenProcessToken.Call(uintptr(currProc),
		uintptr(TOKEN_ADJUST_PRIVILEGES|TOKEN_QUERY),
		uintptr(unsafe.Pointer(&hToken)))
	if ret == 0 {
		return fmt.Errorf("OpenProcessToken failed: %v", err)
	}
	defer syscall.CloseHandle(hToken)

	var luid windows.LUID
	ret2, _, err := procLookupPrivilegeValue.Call(0,
		uintptr(unsafe.Pointer(syscall.StringBytePtr(priv))),
		uintptr(unsafe.Pointer(&luid)))
	if ret2 == 0 {
		return fmt.Errorf("LookupPrivilegeValue failed: %v", err)
	}

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: SE_PRIVILEGE_ENABLED},
		},
	}

	ret3, _, err := procAdjustTokenPrivileges.Call(
		uintptr(hToken),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0)
	if ret3 == 0 {
		return fmt.Errorf("AdjustTokenPrivileges failed: %v", err)
	}
	return nil
}
