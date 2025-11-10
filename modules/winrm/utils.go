package main

import (
	"bytes"
	"context"
	"time"
	"winrm/debugger"
	"winrm/winrm"
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
	endpoint := winrm.NewEndpoint(host, port, tls, insecure, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, username, password)
	if err != nil {
		debugger.Println("Error In New Client", err.Error())
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdoutBuf, stderrBuf bytes.Buffer
	r, err := client.RunWithContext(ctx, command, &stdoutBuf, &stderrBuf)
	debugger.Println("Result ", r)
	if err != nil {
		debugger.Println("Error with RunwithContext", err.Error())
		return nil, err
	}
	debugger.Println("Stdout: ", stdoutBuf.Bytes())
	debugger.Println("Stderr: ", stderrBuf.Bytes())

	stdout = append(stdoutBuf.Bytes(), stderrBuf.Bytes()...)
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


	connectTimeout, err := time.ParseDuration("5s")
	endpoint := winrm.NewEndpoint(host, port, tls, insecure, nil, nil, nil, connectTimeout)

	encryption, err := winrm.NewEncryption("ntlm")
	params := winrm.DefaultParameters
	params.TransportDecorator = func() winrm.Transporter { return encryption }
	debugger.Println("TransportDecorator is -----------------------------------------> ")
	debugger.Println(params.TransportDecorator)

	client, err := winrm.NewClientWithParameters(endpoint, username, password, params)
	if err != nil {
		debugger.Println("Error In New Client", err.Error())
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var stdoutBuf, stderrBuf bytes.Buffer
	r, err := client.RunWithContext(ctx, command, &stdoutBuf, &stderrBuf)
	debugger.Println("Result ", r)
	if err != nil {
		debugger.Println("Error with RunwithContext", err.Error())
		return nil, err
	}
	debugger.Println("Stdout: ", stdoutBuf.Bytes())
	debugger.Println("Stderr: ", stderrBuf.Bytes())

	stdout = append(stdoutBuf.Bytes(), stderrBuf.Bytes()...)
	return stdout, nil
}
