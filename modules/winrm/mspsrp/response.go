package mspsrp

import (
	"encoding/xml"
	"fmt"
)

// SOAP envelope structures for parsing WinRM responses

// CreateShellResponse represents the response from a shell creation request
type CreateShellResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		Shell struct {
			ShellId string `xml:"ShellId"`
		} `xml:"Shell"`
	} `xml:"Body"`
}

// CommandResponse represents the response from a command execution request
type CommandResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		CommandResponse struct {
			CommandId string `xml:"CommandId"`
		} `xml:"CommandResponse"`
	} `xml:"Body"`
}

// ReceiveResponse represents the response from a receive request
type ReceiveResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		ReceiveResponse struct {
			Stream []struct {
				Name      string `xml:"Name,attr"`
				CommandId string `xml:"CommandId,attr"`
				Content   string `xml:",chardata"`
			} `xml:"Stream"`
			CommandState struct {
				CommandId string `xml:"CommandId,attr"`
				State     string `xml:"State,attr"`
				ExitCode  int    `xml:"ExitCode"`
			} `xml:"CommandState"`
		} `xml:"ReceiveResponse"`
	} `xml:"Body"`
}

// ParseShellID extracts the ShellId from a CreateResponse
func ParseShellID(responseXML []byte) (string, error) {
	var response CreateShellResponse
	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Body.Shell.ShellId == "" {
		return "", fmt.Errorf("ShellId not found in response")
	}

	return response.Body.Shell.ShellId, nil
}

// ParseCommandID extracts the CommandId from a CommandResponse
func ParseCommandID(responseXML []byte) (string, error) {
	var response CommandResponse
	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Body.CommandResponse.CommandId == "" {
		return "", fmt.Errorf("CommandId not found in response")
	}

	return response.Body.CommandResponse.CommandId, nil
}

// ParseReceiveOutput extracts the output streams from a ReceiveResponse
func ParseReceiveOutput(responseXML []byte) (stdout []string, stderr []string, exitCode int, done bool, err error) {
	var response ReceiveResponse
	if err := xml.Unmarshal(responseXML, &response); err != nil {
		return nil, nil, 0, false, fmt.Errorf("failed to parse response: %w", err)
	}

	// Collect all stdout and stderr streams (there can be multiple)
	for _, stream := range response.Body.ReceiveResponse.Stream {
		switch stream.Name {
		case "stdout":
			stdout = append(stdout, stream.Content)
		case "stderr":
			stderr = append(stderr, stream.Content)
		}
	}

	// Check if command is done
	state := response.Body.ReceiveResponse.CommandState.State
	done = state == "http://schemas.microsoft.com/wbem/wsman/1/windows/shell/CommandState/Done"
	exitCode = response.Body.ReceiveResponse.CommandState.ExitCode

	return stdout, stderr, exitCode, done, nil
}

