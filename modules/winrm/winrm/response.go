package gowinrm

import (
	"io"
	"winrm/debugger"
	"winrm/mspsrp"
	"winrm/pwshxml"
)

type ExecuteCommandError struct {
	Inner error
	Body  string
}

func (e *ExecuteCommandError) Error() string {
	if e.Inner == nil {
		return "error"
	}

	return e.Inner.Error()
}

func (e *ExecuteCommandError) Is(err error) bool {
	_, ok := err.(*ExecuteCommandError)
	return ok
}

func (b *ExecuteCommandError) Unwrap() error {
	return b.Inner
}

// ParseSlurpOutputErrResponse ParseSlurpOutputErrResponse
func ParseSlurpOutputErrResponse(response string, stdout, stderr io.Writer) (bool, int, error) {

	stdoutByte, stderrByte, exitcode, done, err := mspsrp.ParseReceiveOutput([]byte(response))
	debugger.Println("Stdout Streams: ", stdoutByte)
	debugger.Println("Stderr Streams: ", stderrByte)

	cleanStdout, err := pwshxml.DecodeStreams(stdoutByte)
	debugger.Println("Error While Stdout Decoding ", err)
	for _, line := range cleanStdout {
		debugger.Println("Stdout: ", line)
		stdout.Write([]byte(line))
		stdout.Write([]byte("\n"))
	}

	cleanStderr, err := pwshxml.DecodeStreams(stderrByte)
	debugger.Println("Error While Stderr Decoding ", err)
	for _, line := range cleanStderr {
		debugger.Println("Stderr: ", line)
		stderr.Write([]byte(line))
		stderr.Write([]byte("\n"))
	}

	return done, exitcode, err
}
