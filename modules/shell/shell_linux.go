package main

import (
	"bytes"
	"io"
	"os/exec"
	"time"

    "github.com/creack/pty"
)

type Shell struct {
	Pid        int
	StdinPipe  io.WriteCloser
	StdoutPipe io.ReadCloser
	StderrPipe io.ReadCloser
	StdoutBuf  *bytes.Buffer
	StderrBuf  *bytes.Buffer
}

var LIST_SHELL []*Shell

func StartShell() (*Shell, error) {
    shellPath := "bash"
	c := exec.Command(shellPath)

    ptyFile, err := pty.Start(c)
	if err != nil {
		return nil, err
	}

	s := &Shell{
		Pid:        c.Process.Pid,
		StdinPipe:  ptyFile,
		StdoutPipe: ptyFile,
		StderrPipe: ptyFile,
		StdoutBuf:  bytes.NewBuffer(nil),
		StderrBuf:  bytes.NewBuffer(nil),
	}
	time.Sleep(100 * time.Millisecond)
    LIST_SHELL = append(LIST_SHELL, s)

	go s.continuousRead(s.StdoutPipe, s.StdoutBuf)
	go s.continuousRead(s.StderrPipe, s.StderrBuf)
	return s, nil
}

func (s *Shell) continuousRead(pipe io.Reader, buf *bytes.Buffer) {
	ioBuffer := make([]byte, 4096)
	for {
		n, _ := pipe.Read(ioBuffer)
		if n > 0 {
			buf.Write(ioBuffer[:n])
		}
	}
}

func (s *Shell) ReadOut() (stdout []byte, err error) {
    time.Sleep(100 * time.Millisecond)
    res := s.StdoutBuf.Bytes()
    s.StdoutBuf.Reset()
	return res, nil
}

func (s *Shell) ReadErr() (stderr []byte, err error) {
    time.Sleep(100 * time.Millisecond)
    res := s.StderrBuf.Bytes()
    s.StderrBuf.Reset()
	return res, nil
}

func (s *Shell) Write(input string) error {
	_, err := s.StdinPipe.Write([]byte(input + "\n"))
	if err != nil {
		return err
	}
	return nil
}
