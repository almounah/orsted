package main

import (
	"fmt"
	"strconv"
	"strings"
	"winrm/debugger"
)


func ExecuteWinrmCommandLocalUsernamePassword(host string, port int, tls bool, insecure bool, username string, password string, command string) (stdout []byte, err error) {
	debugger.Println("AuthType Local")
	debugger.Println("Host:", host)
	debugger.Println("Port:", port)
	debugger.Println("TLS:", tls)
	debugger.Println("Insecure:", insecure)
	debugger.Println("Username:", username)
	debugger.Println("Password:", password)
	debugger.Println("Command:", command)

	return stdout, nil
}

func ExecuteWinrmCommandNTLMUsernamePassword(host string, port int, tls bool, insecure bool, username string, password string, command string) (stdout []byte, err error) {
	debugger.Println("AuthType NTLM / Domain")
	debugger.Println("Host:", host)
	debugger.Println("Port:", port)
	debugger.Println("TLS:", tls)
	debugger.Println("Insecure:", insecure)
	debugger.Println("Username:", username)
	debugger.Println("Password:", password)
	debugger.Println("Command:", command)

	shell := NewWinRMShell(
		"http://" + host + ":" + strconv.Itoa(port) + "/wsman",
		"",
		username,
		password,
	)

	// Step 1: Create the PowerShell shell (performs NTLM handshake)
	if err = shell.CreateShell(); err != nil {
		fmt.Printf("Error creating shell: %v\n", err)
		return []byte(err.Error()), err
	}

	// Step 2: Execute a command (uses encrypted channel)
	commandID, err := shell.ExecuteCommand("Get-ChildItem")
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return []byte(err.Error()), err
	}

	// Step 3: Receive output (uses encrypted channel)
	stdoutStr, err := shell.ReceiveOutput(commandID)
	if err != nil {
		fmt.Printf("Error receiving output: %v\n", err)
		return []byte(err.Error()), err
	}

	s := strings.Join(stdoutStr, "\n") // concatenate all strings
    stdout = []byte(s)


	return stdout, nil
}
