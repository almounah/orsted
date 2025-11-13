package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
	"winrm/debugger"
	"winrm/ntlmssp"
	gowinrm "winrm/winrm"
)

func isLMNT(s string) bool {
	re := regexp.MustCompile(`^[0-9a-fA-F]{32}:[0-9a-fA-F]{32}$`)
	return re.MatchString(s)
}

func ExecuteWinrmCommandNTLMUsernamePassword(host string, port int, tls bool, insecure bool, username string, password string, hash string, command string) (stdout []byte, err error) {
	debugger.Println("AuthType NTLM / Domain")
	debugger.Println("Host:", host)
	debugger.Println("Port:", port)
	debugger.Println("TLS:", tls)
	debugger.Println("Insecure:", insecure)
	debugger.Println("Username:", username)
	debugger.Println("Password:", password)
	debugger.Println("Domain", password)
	debugger.Println("Hash", hash)
	debugger.Println("Command:", command)

	if hash != "" && password != "" {
		err := fmt.Errorf("Cannot specify Hash and password. Only One")
		return []byte(err.Error()), err
	}

	if password != "" {
		lmByte, err := ntlmssp.LmowfV1(password)
		if err != nil {
			return []byte(err.Error()), err
		}
		ntByte, err := ntlmssp.NtowfV1(password)
		if err != nil {
			return []byte(err.Error()), err
		}
		lm := hex.EncodeToString(lmByte)
		nt := hex.EncodeToString(ntByte)
		hash = fmt.Sprintf("%s:%s", lm, nt)
	}

	if !strings.Contains(hash,":") {
		hash = "00000000000000000000000000000000:"+hash
	} else {
		if !isLMNT(hash) {
			err := fmt.Errorf("Wrong hash format")
			return []byte(err.Error()), err
		}
	}

	connectTimeout, err := time.ParseDuration("5s")

	endpoint := gowinrm.NewEndpoint(host, port, tls, insecure, nil, nil, nil, connectTimeout)

	encryption, err := gowinrm.NewEncryption("ntlm")

	params := gowinrm.DefaultParameters


	params.TransportDecorator = func() gowinrm.Transporter { return encryption }


	client, err := gowinrm.NewClientWithParameters(endpoint, username, hash, params)

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
