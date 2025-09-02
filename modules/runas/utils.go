package main

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/sys/windows"
)

var Advapi32 = windows.NewLazySystemDLL("Advapi32.dll")

const (
	LOGON_WITH_PROFILE        uint32 = 0x1
	LOGON_NETCREDENTIALS_ONLY uint32 = 0x2
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

func GetTokenUsername(token windows.Token) (username string, err error) {
	user, err := token.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("there was an error calling GetTokenUser(): %s", err)
	}

	account, domain, _, err := user.User.Sid.LookupAccount("")
	if err != nil {
		return "", fmt.Errorf("there was an error calling SID.LookupAccount(): %s", err)
	}

	username = fmt.Sprintf("%s\\%s", domain, account)
	return
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

func CreateProcessWithLogon(lpUsername *uint16, lpDomain *uint16, lpPassword *uint16, dwLogonFlags uint32, lpApplicationName *uint16, lpCommandLine *uint16, dwCreationFlags uint32, lpEnvironment uintptr, lpCurrentDirectory *uint16, lpStartupInfo *windows.StartupInfo, lpProcessInformation *windows.ProcessInformation) error {
	CreateProcessWithLogonW := Advapi32.NewProc("CreateProcessWithLogonW")

	// Parse optional arguments
	var domain uintptr
	if *lpDomain == 0 {
		domain = 0
	} else {
		domain = uintptr(unsafe.Pointer(lpDomain))
	}

	var applicationName uintptr
	if *lpApplicationName == 0 {
		applicationName = 0
	} else {
		applicationName = uintptr(unsafe.Pointer(lpApplicationName))
	}

	var commandLine uintptr
	if *lpCommandLine == 0 {
		commandLine = 0
	} else {
		commandLine = uintptr(unsafe.Pointer(lpCommandLine))
	}

	var currentDirectory uintptr
	if *lpCurrentDirectory == 0 {
		currentDirectory = 0
	} else {
		currentDirectory = uintptr(unsafe.Pointer(lpCurrentDirectory))
	}

	// BOOL CreateProcessWithLogonW(
	//  [in]                LPCWSTR               lpUsername,
	//  [in, optional]      LPCWSTR               lpDomain,
	//  [in]                LPCWSTR               lpPassword,
	//  [in]                DWORD                 dwLogonFlags,
	//  [in, optional]      LPCWSTR               lpApplicationName, The function does not use the search path
	//  [in, out, optional] LPWSTR                lpCommandLine, The maximum length of this string is 1024 characters.
	//  [in]                DWORD                 dwCreationFlags,
	//  [in, optional]      LPVOID                lpEnvironment,
	//  [in, optional]      LPCWSTR               lpCurrentDirectory,
	//  [in]                LPSTARTUPINFOW        lpStartupInfo,
	//  [out]               LPPROCESS_INFORMATION lpProcessInformation
	//);
	ret, _, err := CreateProcessWithLogonW.Call(
		uintptr(unsafe.Pointer(lpUsername)),
		domain,
		uintptr(unsafe.Pointer(lpPassword)),
		uintptr(dwLogonFlags),
		applicationName,
		commandLine,
		uintptr(dwCreationFlags),
		lpEnvironment,
		currentDirectory,
		uintptr(unsafe.Pointer(lpStartupInfo)),
		uintptr(unsafe.Pointer(lpProcessInformation)),
	)
	if err != syscall.Errno(0) || ret == 0 {
		return fmt.Errorf("there was an error calling CreateProcessWithLogon with return code %d: %s", ret, err)
	}
	return nil
}

func CreateProcessWithLogonWrapper(username string, domain string, password string, application string, args string, logon uint32, hide bool, background bool) (stdout string, stderr string) {
	if username == "" {
		stderr = "a username must be provided for the CreateProcessWithLogon call"
		return
	}

	if password == "" {
		stderr = "a password must be provided for the CreateProcessWithLogon call"
		return
	}

	if application == "" {
		stderr = "an application must be provided for the CreateProcessWithLogon call"
		return
	}

	// Check for UPN format (e.g., rastley@acme.com)
	if strings.Contains(username, "@") {
		temp := strings.Split(username, "@")
		username = temp[0]
		domain = temp[1]
	}

	// Check for domain format (e.g., ACME\rastley)
	if strings.Contains(username, "\\") {
		temp := strings.Split(username, "\\")
		username = temp[1]
		domain = temp[0]
	}

	// Check for an empty or missing domain; used with local user accounts
	if domain == "" {
		domain = "."
	}

	// Convert the username to a LPCWSTR
	lpUsername, err := syscall.UTF16PtrFromString(username)
	if err != nil {
		stderr = fmt.Sprintf("there was an error converting the username \"%s\" to LPCWSTR: %s", username, err)
		return
	}

	// Convert the domain to a LPCWSTR
	lpDomain, err := syscall.UTF16PtrFromString(domain)
	if err != nil {
		stderr = fmt.Sprintf("there was an error converting the domain \"%s\" to LPCWSTR: %s", domain, err)
		return
	}

	// Convert the password to a LPCWSTR
	lpPassword, err := syscall.UTF16PtrFromString(password)
	if err != nil {
		stderr = fmt.Sprintf("there was an error converting the password \"%s\" to LPCWSTR: %s", password, err)
		return
	}

	// Search PATH environment variable to retrieve the application's absolute path
	application, err = exec.LookPath(application)
	if err != nil {
		stderr = fmt.Sprintf("there was an error resolving the absolute path for %s: %s", application, err)
		return
	}

	// Convert the application to a LPCWSTR
	lpApplicationName, err := syscall.UTF16PtrFromString(application)
	if err != nil {
		stderr = fmt.Sprintf("there was an error converting the application name \"%s\" to LPCWSTR: %s", application, err)
		return
	}

	// Convert the program to a LPCWSTR
	lpCommandLine, err := syscall.UTF16PtrFromString(args)
	if err != nil {
		stderr = fmt.Sprintf("there was an error converting the application arguments \"%s\" to LPCWSTR: %s", args, err)
		return
	}

	// Setup pipes to retrieve output
	stdInRead, _, stdOutRead, stdOutWrite, stdErrRead, stdErrWrite, err := CreateAnonymousPipes()
	if err != nil {
		stderr = fmt.Sprintf("there was an error creating anonymous pipes to collect output: %s", err)
		return
	}

	lpCurrentDirectory := uint16(0)
	lpStartupInfo := windows.StartupInfo{}
	if !background {
		lpStartupInfo = windows.StartupInfo{
			StdInput:  stdInRead,
			StdOutput: stdOutWrite,
			StdErr:    stdErrWrite,
			Flags:     windows.STARTF_USESTDHANDLES,
		}
	}
	if hide {
		lpStartupInfo.Flags = windows.STARTF_USESTDHANDLES | windows.STARTF_USESHOWWINDOW
		lpStartupInfo.ShowWindow = windows.SW_HIDE
	}
	lpProcessInformation := windows.ProcessInformation{}

	err = CreateProcessWithLogon(
		lpUsername,
		lpDomain,
		lpPassword,
		logon,
		lpApplicationName,
		lpCommandLine,
		0,
		0,
		&lpCurrentDirectory,
		&lpStartupInfo,
		&lpProcessInformation,
	)

	if err != nil {
		stderr += err.Error()
		return
	}

	stdout += fmt.Sprintf("Created %s %s process with an ID of %d\n", application, args, lpProcessInformation.ProcessId)

	// Close the "write" pipe handles
	err = ClosePipes(0, 0, 0, stdOutWrite, 0, stdErrWrite)
	if err != nil {
		stderr = err.Error()
		return
	}

	// Read from the pipes
	var out string
	_, out, stderr, err = ReadPipes(0, stdOutRead, stdErrRead)
	if err != nil {
		stderr += err.Error()
		return
	}
	stdout += out

	// Close the "read" pipe handles
	err = ClosePipes(stdInRead, 0, stdOutRead, 0, stdErrRead, 0)
	if err != nil {
		stderr += err.Error()
		return
	}
	return
}

func executeCommandWithAttributes(name string, args []string, attr *syscall.SysProcAttr) (stdout string, stderr string) {
	application, err := exec.LookPath(name)
	if err != nil {
		stderr = fmt.Sprintf("there was an error resolving the absolute path for %s: %s", application, err)
		return
	}

	// #nosec G204 -- Subprocess must be launched with a variable
	cmd := exec.Command(application, args...)
	cmd.SysProcAttr = attr

	out, err := cmd.CombinedOutput()
	if cmd.Process != nil {
		stdout = fmt.Sprintf("Created %s process with an ID of %d\n", application, cmd.Process.Pid)
	}

	// Convert the output to a string
	if utf8.Valid(out) {
		stdout += string(out)
	} else {
		s, e := DecodeString(out)
		if e != nil {
			stderr = fmt.Sprintf("%s\n", e)
		} else {
			stdout += s
		}
	}

	if err != nil {
		stderr += err.Error()
	}

	return stdout, stderr
}
