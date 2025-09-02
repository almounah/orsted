package main

/*
#include <stdlib.h>
*/
import "C"
import (
	"bytes"
	"os/exec"

	"github.com/desertbit/go-shlex"
)

//export Run
func Run(command string) (stdout []byte, err error) {
	
	parts, err := shlex.Split(command, false)
	if err != nil {
		// handle error
	}
	cmd := exec.Command(parts[0], parts[1:]...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()

	// Allocate and populate stdout
	stdoutBytes := outBuf.Bytes()
	sterrBytes := errBuf.Bytes()

	stdout = append(stdoutBytes, sterrBytes...)
	return stdout, err
}

