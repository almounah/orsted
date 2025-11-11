package pwshxml

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// CreateCommandArguments generates the base64-encoded command arguments for executing a command
// This is used in the <rsp:Arguments> element when sending a Command request
func CreateCommandArguments(command string, runspaceID, pipelineID string) (string, error) {
	runspaceUUID, err := uuid.Parse(runspaceID)
	if err != nil {
		return "", fmt.Errorf("invalid runspace ID: %w", err)
	}

	pipelineUUID, err := uuid.Parse(pipelineID)
	if err != nil {
		return "", fmt.Errorf("invalid pipeline ID: %w", err)
	}

	// Build the CREATE_PIPELINE message with Invoke-Expression and Out-String
	pipelineXML := buildCreatePipelineXML(command)

	var fragments bytes.Buffer
	objectID := uint64(1)

	if err := writeFragment(&fragments, CREATE_PIPELINE, pipelineXML, runspaceUUID, pipelineUUID, objectID); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(fragments.Bytes()), nil
}

// buildCreatePipelineXML builds the CREATE_PIPELINE XML for executing a command
func buildCreatePipelineXML(command string) string {
	var b strings.Builder
	b.WriteString("<Obj>")
	b.WriteString("<MS>")
	
	// NoInput
	b.WriteString(`<B N="NoInput">true</B>`)
	
	// AddToHistory
	b.WriteString(`<B N="AddToHistory">false</B>`)
	
	// IsNested
	b.WriteString(`<B N="IsNested">false</B>`)
	
	// ApartmentState = 2 (Unknown)
	b.WriteString(`<Obj N="ApartmentState"><I32>2</I32></Obj>`)
	
	// RemoteStreamOptions = 15 (AddInvocationInfo)
	b.WriteString(`<Obj N="RemoteStreamOptions"><I32>15</I32></Obj>`)
	
	// HostInfo
	b.WriteString(`<Obj N="HostInfo">`)
	b.WriteString("<MS>")
	b.WriteString(`<B N="_isHostNull">true</B>`)
	b.WriteString(`<B N="_isHostUINull">true</B>`)
	b.WriteString(`<B N="_isHostRawUINull">true</B>`)
	b.WriteString(`<B N="_useRunspaceHost">true</B>`)
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	
	// PowerShell object with commands
	b.WriteString(`<Obj N="PowerShell">`)
	b.WriteString("<MS>")
	b.WriteString(`<B N="IsNested">false</B>`)
	b.WriteString(`<B N="RedirectShellErrorOutputPipe">false</B>`)
	b.WriteString(`<Nil N="ExtraCmds"/>`)
	b.WriteString(`<Nil N="History"/>`)
	
	// Commands list
	b.WriteString(`<Obj N="Cmds">`)
	b.WriteString("<LST>")
	
	// First command: Invoke-Expression with the actual command
	b.WriteString(buildCommandXML("Invoke-Expression", map[string]string{
		"Command": command,
	}))
	
	// Second command: Out-String with Stream parameter
	b.WriteString(buildCommandXML("Out-String", map[string]string{
		"Stream": "", // Nil value
	}))
	
	b.WriteString("</LST>")
	b.WriteString("</Obj>")
	
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	return b.String()
}

// buildCommandXML builds a single command object with arguments
func buildCommandXML(cmdName string, args map[string]string) string {
	var b strings.Builder
	b.WriteString("<Obj>")
	b.WriteString("<MS>")
	
	// Command name
	b.WriteString(`<S N="Cmd">`)
	b.WriteString(xmlEscape(cmdName))
	b.WriteString("</S>")
	
	// Arguments
	b.WriteString(`<Obj N="Args">`)
	b.WriteString("<LST>")
	for name, value := range args {
		b.WriteString("<Obj>")
		b.WriteString("<MS>")
		b.WriteString(`<S N="N">`)
		b.WriteString(xmlEscape(name))
		b.WriteString("</S>")
		
		if value == "" {
			b.WriteString(`<Nil N="V"/>`)
		} else {
			b.WriteString(`<S N="V">`)
			b.WriteString(xmlEscape(value))
			b.WriteString("</S>")
		}
		b.WriteString("</MS>")
		b.WriteString("</Obj>")
	}
	b.WriteString("</LST>")
	b.WriteString("</Obj>")
	
	// IsScript
	b.WriteString(`<B N="IsScript">false</B>`)
	
	// UseLocalScope
	b.WriteString(`<Nil N="UseLocalScope"/>`)
	
	// Pipeline result types (all set to 0 = None/Default)
	b.WriteString(`<Obj N="MergeMyResult"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeToResult"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergePreviousResults"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeError"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeWarning"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeVerbose"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeDebug"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="MergeInformation"><I32>0</I32></Obj>`)
	
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	return b.String()
}

// xmlEscape escapes special XML characters
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

