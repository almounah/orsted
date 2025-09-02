package main

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	LOGON32_LOGON_INTERACTIVE       uint32 = 2
	LOGON32_LOGON_NETWORK           uint32 = 3
	LOGON32_LOGON_BATCH             uint32 = 4
	LOGON32_LOGON_SERVICE           uint32 = 5
	LOGON32_LOGON_UNLOCK            uint32 = 7
	LOGON32_LOGON_NETWORK_CLEARTEXT uint32 = 8
	LOGON32_LOGON_NEW_CREDENTIALS   uint32 = 9
)

// LOGON32_PROVIDER_ constants
// The logon provider
const (
	LOGON32_PROVIDER_DEFAULT uint32 = iota
	LOGON32_PROVIDER_WINNT35
	LOGON32_PROVIDER_WINNT40
	LOGON32_PROVIDER_WINNT50
	LOGON32_PROVIDER_VIRTUAL
)

func maketoken(username, password, logonType string) ([]byte, error) {
	logTypeInt, err := strconv.Atoi(logonType)
	if err != nil {
		return []byte("Invalid logon type"), err
	}
	if logTypeInt != int(LOGON32_LOGON_INTERACTIVE) && 
	logTypeInt != int(LOGON32_LOGON_NETWORK) && 
	logTypeInt != int(LOGON32_LOGON_BATCH) &&
	logTypeInt != int(LOGON32_LOGON_SERVICE) &&
	logTypeInt != int(LOGON32_LOGON_UNLOCK) && 
	logTypeInt != int(LOGON32_LOGON_NETWORK_CLEARTEXT) &&
	logTypeInt != int(LOGON32_LOGON_NEW_CREDENTIALS) {
		return []byte("Invalid logon type"), err
	}
	// Make token
	token, err := LogonUser(username, password, "", uint32(logTypeInt), LOGON32_PROVIDER_DEFAULT)
	if err != nil {
		return []byte(""), err
	}

	// Get Token Stats
	stats, err := GetTokenStats(token)
	if err != nil {
		return []byte(""), err
	}

	res := fmt.Sprintf("Successfully created a Windows access token for %s with a logon ID of 0x%X. Token ID 0x%X", username, stats.AuthenticationId.LowPart, stats.TokenId.LowPart)
	fmt.Println("Applying token in thread ---> ")
	fmt.Println(windows.GetCurrentThreadId())
	err = ApplyToken(token)
	if err != nil {
		fmt.Println("Error Applying Token ", err.Error())
		return []byte(res), err
	}

	return []byte(res), nil
}

func LogonUser(user string, password string, domain string, logonType uint32, logonProvider uint32) (hToken windows.Token, err error) {
	if user == "" {
		err = fmt.Errorf("a username must be provided for the LogonUser call")
		return
	}

	if password == "" {
		err = fmt.Errorf("a password must be provided for the LogonUser call")
		return
	}

	if logonType <= 0 {
		err = fmt.Errorf("an invalid logonType was provided to the LogonUser call: %d", logonType)
		return
	}

	// Check for UPN format (e.g., rastley@acme.com)
	if strings.Contains(user, "@") {
		temp := strings.Split(user, "@")
		user = temp[0]
		domain = temp[1]
	}

	// Check for domain format (e.g., ACME\rastley)
	if strings.Contains(user, "\\") {
		temp := strings.Split(user, "\\")
		user = temp[1]
		domain = temp[0]
	}

	// Check for an empty or missing domain; used with local user accounts
	if domain == "" {
		domain = "."
	}

	// Convert username to LPCWSTR
	pUser, err := windows.UTF16PtrFromString(user)
	if err != nil {
		err = fmt.Errorf("there was an error converting the username \"%s\" to LPCWSTR: %s", user, err)
		return
	}

	// Convert the domain to LPCWSTR
	pDomain, err := windows.UTF16PtrFromString(domain)
	if err != nil {
		err = fmt.Errorf("there was an error converting the domain \"%s\" to LPCWSTR: %s", domain, err)
		return
	}

	// Convert the password to LPCWSTR
	pPassword, err := windows.UTF16PtrFromString(password)
	if err != nil {
		err = fmt.Errorf("there was an error converting the password \"%s\" to LPCWSTR: %s", password, err)
		return
	}

	token, err := advapi32LogonUser(pUser, pDomain, pPassword, logonType, logonProvider)
	if err != nil {
		return
	}

	// Convert *unsafe.Pointer to windows.Token
	// windows.Token -> windows.Handle -> uintptr
	hToken = (windows.Token)(*token)
	return
}

// LogonUser attempts to log a user on to the local computer.
// The local computer is the computer from which LogonUser was called. You cannot use LogonUser to log on to a remote computer.
// You specify the user with a user name and domain and authenticate the user with a plaintext password.
// If the function succeeds, you receive a handle to a token that represents the logged-on user.
// You can then use this token handle to impersonate the specified user or, in most cases, to create a process that runs in the context of the specified user.
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-logonuserw
func advapi32LogonUser(lpszUsername *uint16, lpszDomain *uint16, lpszPassword *uint16, dwLogonType uint32, dwLogonProvider uint32) (token *unsafe.Pointer, err error) {
	// The LogonUser function was not available in the golang.org/x/sys/windows package at the time of writing
	LogonUserW := Advapi32.NewProc("LogonUserW")

	// BOOL LogonUserW(
	//  [in]           LPCWSTR lpszUsername,
	//  [in, optional] LPCWSTR lpszDomain,
	//  [in, optional] LPCWSTR lpszPassword,
	//  [in]           DWORD   dwLogonType,
	//  [in]           DWORD   dwLogonProvider,
	//  [out]          PHANDLE phToken
	//);

	var phToken unsafe.Pointer

	_, _, err = LogonUserW.Call(
		uintptr(unsafe.Pointer(lpszUsername)),
		uintptr(unsafe.Pointer(lpszDomain)),
		uintptr(unsafe.Pointer(lpszPassword)),
		uintptr(dwLogonType),
		uintptr(dwLogonProvider),
		uintptr(unsafe.Pointer(&phToken)),
	)
	if err != windows.Errno(0) {
		err = fmt.Errorf("there was an error calling advapi32!LogonUserW: %s", err)
		return
	}
	return &phToken, nil
}
