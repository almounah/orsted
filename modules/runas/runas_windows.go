package main

import (
	"fmt"
	"errors"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

// RunAs creates a new process as the provided user
func RunAs(username string, password string, domain string, application []string, background bool) ([]byte, error) {

	// Username, Password, Application, Arguments
	var arguments string
	if len(application) > 1 {
		arguments = strings.Join(application[0:], " ")
	}

	applicationString := application[0]
	Println(fmt.Sprintf("%s %s %s %s", username, password, applicationString, arguments))

	// Determine if running as SYSTEM
	u, err := GetTokenUsername(windows.GetCurrentProcessToken())
	if err != nil {
		results := err.Error()
		return []byte(results), err
	}

	// If we are running as SYSTEM, we can't call CreateProcess, must call LogonUserA -> CreateProcessAsUserA/CreateProcessWithTokenW
	if u == "NT AUTHORITY\\SYSTEM" {
		hToken, err2 := LogonUser(username, password, "", LOGON32_LOGON_INTERACTIVE, LOGON32_PROVIDER_DEFAULT)
		if err2 != nil {
			results := err2.Error()
			return []byte(results), err2
		}
		//results.Stdout, results.Stderr = tokens.CreateProcessWithToken(hToken, application, strings.Split(arguments, " "))
		var args []string
		if len(application) > 1 {
			args = application[1:]
		}

		attr := &windows.SysProcAttr{
			HideWindow: true,
			Token:      syscall.Token(hToken),
		}
		stdout, stderr := executeCommandWithAttributes(applicationString, args, attr)
		return []byte(stdout), errors.New((string(stderr)))
	}

	stdout, stderr := CreateProcessWithLogonWrapper(username, domain, password, applicationString, arguments, LOGON_WITH_PROFILE, true, background)
	combinedOutput := append([]byte(stdout), []byte(stderr)...)

	return []byte(combinedOutput), errors.New((string(stderr)))
}
