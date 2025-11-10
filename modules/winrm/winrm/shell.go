package winrm

import (
	"context"
	"winrm/debugger"
)

// Shell is the local view of a WinRM Shell of a given Client
type Shell struct {
	client    *Client
	id        string
	sessionId string
	runspaceId string
	pipelineId string
}

// Execute command on the given Shell, returning either an error or a Command
//
// Deprecated: user ExecuteWithContext
func (s *Shell) Execute(command string, arguments ...string) (*Command, error) {
	return s.ExecuteWithContext(context.Background(), command, arguments...)
}

// ExecuteWithContext command on the given Shell, returning either an error or a Command
func (s *Shell) ExecuteWithContext(ctx context.Context, command string, arguments ...string) (*Command, error) {
	request := NewExecutePowerShellCommandRequest(s.client.url, s.id, s.sessionId, s.runspaceId, s.pipelineId, command, arguments, &s.client.Parameters)
	debugger.Println("Request Execute PowerShell Command")
	debugger.Println("---------------------------------")
	debugger.Println(request.String())
	debugger.Println("---------------------------------")
	defer request.Free()

	response, err := s.client.sendRequest(request)
	debugger.Println()
	debugger.Println()
	debugger.Println("Response Execute PowerShell Command")
	debugger.Println("---------------------------------")
	debugger.Println(response)
	debugger.Println("---------------------------------")
	if err != nil {
		return nil, err
	}

	commandID, err := ParseExecuteCommandResponse(response)
	if err != nil {
		return nil, err
	}

	cmd := newCommand(ctx, s, commandID)

	return cmd, nil
}

// Close will terminate this shell. No commands can be issued once the shell is closed.
func (s *Shell) Close() error {
	request := NewDeleteShellRequest(s.client.url, s.id, &s.client.Parameters)
	defer request.Free()

	_, err := s.client.sendRequest(request)
	return err
}
