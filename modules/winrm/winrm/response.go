package gowinrm

import (
	"fmt"
	"io"
	"winrm/mspsrp"

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



func newExecuteCommandError(response string, format string, args ...interface{}) *ExecuteCommandError {
	return &ExecuteCommandError{fmt.Errorf(format, args...), response}
}



// ParseSlurpOutputErrResponse ParseSlurpOutputErrResponse
func ParseSlurpOutputErrResponse(response string, stdout, stderr io.Writer) (bool, int, error) {

	stdoutByte, stderrByte, exitcode, done, err := mspsrp.ParseReceiveOutput([]byte(response))
	for _, node := range stdoutByte {
		stdout.Write([]byte(node))
	}

	for _, node := range stderrByte {
		stderr.Write([]byte(node))
	}

	return done, exitcode, err
}

