package main

import (
	"bytes"
	"context"
	"time"
	"winrm/debugger"
	gowinrm "winrm/winrm"
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

	connectTimeout, err := time.ParseDuration("5s")

	endpoint := gowinrm.NewEndpoint(host, port, tls, insecure, nil, nil, nil, connectTimeout)

	encryption, err := gowinrm.NewEncryption("ntlm")

	params := gowinrm.DefaultParameters


	params.TransportDecorator = func() gowinrm.Transporter { return encryption }


	client, err := gowinrm.NewClientWithParameters(endpoint, username, password, params)

	if err != nil {

		debugger.Println("Error In New Client", err.Error())

		return nil, err

	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stdoutBuf, stderrBuf bytes.Buffer
	_, err = client.RunWithContext(ctx, command, &stdoutBuf, &stderrBuf)
	if err != nil {
		debugger.Println("Error with RunwithContext", err.Error())
		return nil, err
	}

	stdout = append(stdoutBuf.Bytes(), stderrBuf.Bytes()...)

	return stdout, nil
}
