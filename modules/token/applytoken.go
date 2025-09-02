package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

func ApplyToken(Token windows.Token) error {

	// Apply the token to this process thread
	return ImpersonateLoggedOnUser(Token)
}

//func ImpersonateLoggedOnUser(hToken windows.Token) (err error) {
//	impersonateLoggedOnUser := Advapi32.NewProc("ImpersonateLoggedOnUser")
//
//	// BOOL ImpersonateLoggedOnUser(
//	//  [in] HANDLE hToken
//	//);
//	_, _, err = impersonateLoggedOnUser.Call(uintptr(hToken))
//	if err != windows.Errno(0) {
//		err = fmt.Errorf("there was an error calling ImpersonateLoggedOnUser: %s", err)
//		return
//	}
//	err = nil
//	return
//}

const (
    CurrentThread windows.Handle = ^windows.Handle(1) + 1 // -2 as HANDLE
)
func ImpersonateLoggedOnUser(hToken windows.Token) (err error) {
	newToken, err := DuplicateToPrimary(hToken)
	if err != nil {
		fmt.Println("Error DuplicateToPrimary: ", err.Error())
	}
	err = NtSetInformationThread(windows.CurrentThread(), 5, unsafe.Pointer(&newToken), uint32(unsafe.Sizeof(newToken)))
	if err != nil {
		fmt.Println("Error NtSetInformationThread: ", err.Error())
	}
	return err
}

func DuplicateToPrimary(hImp windows.Token) (windows.Token, error) {
    var hPrimary windows.Token
    const (
        // Request maximum access
        desiredAccess = windows.MAXIMUM_ALLOWED
        // Token type: primary
        tokenType = windows.TokenImpersonation
        // Impersonation level: SecurityImpersonation (2)
        impLevel = 2
    )

    // Call DuplicateTokenEx
    err := NtDuplicateToken(
        hImp,
        desiredAccess,
        nil,
        impLevel,
        tokenType,
        &hPrimary,
    )
    if err != nil {
        return 0, fmt.Errorf("DuplicateTokenEx failed: %v", err)
    }
    return hPrimary, nil
}
