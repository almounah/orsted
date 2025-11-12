package gowinrm

import (
	"context"
	"fmt"
	"winrm/debugger"
	"winrm/mspsrp"
	"winrm/pwshxml"

	"github.com/google/uuid"
)

// Shell is the local view of a WinRM Shell of a given Client
type Shell struct {
	Client     *Client
	Id         string
	SessionID  string
	RunspaceID string
	PipelineID string
}

// Execute command on the given Shell, returning either an error or a Command
//
// Deprecated: user ExecuteWithContext
func (s *Shell) Execute(command string, arguments ...string) (*Command, error) {
	return s.ExecuteWithContext(context.Background(), command, arguments...)
}

// ExecuteWithContext command on the given Shell, returning either an error or a Command
func (s *Shell) ExecuteWithContext(ctx context.Context, command string, arguments ...string) (*Command, error) {
	// Step 1: Generate command arguments using your function
	commandArgs, err := pwshxml.CreateCommandArguments(command, s.RunspaceID, s.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to create command arguments: %w", err)
	}

	// Step 2: Build the SOAP request
	messageId := uuid.New().String()
	request, err := mspsrp.EnvelopeToString(mspsrp.CreateCommandRequest(commandArgs, s.Id, s.SessionID, messageId))
	if err != nil {
		return nil, fmt.Errorf("failed to create CreateCommand XML: %w", err)
	}
	debugger.Println("Execute Command Request")
	debugger.Println("----------------------------")
	debugger.Println(request)
	debugger.Println("----------------------------")

	response, err := s.Client.sendRequest([]byte(request))
	if err != nil {
		return nil, err
	}
	debugger.Println("Execute Command Response")
	debugger.Println("----------------------------")
	debugger.Println(response)
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")
	debugger.Println("----------------------------")

	commandID, err := mspsrp.ParseCommandID([]byte(response))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CommandID: %w", err)
	}

	cmd := newCommand(ctx, s, commandID)

	return cmd, nil
}

// Close will terminate this shell. No commands can be issued once the shell is closed.
func (s *Shell) Close() error {
	request := NewDeleteShellRequest(s.Client.url, s.Id, &s.Client.Parameters)

	_, err := s.Client.sendRequest([]byte(request))
	return err
}
