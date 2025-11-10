package winrm

import (
	"encoding/base64"

	"winrm/debugger"
	"winrm/winrm/psrp"
	"winrm/winrm/soap"

	"github.com/gofrs/uuid"
)

func genUUID() string {
	id := uuid.Must(uuid.NewV4())
	return "uuid:" + id.String()
}

func defaultHeaders(message *soap.SoapMessage, url string, params *Parameters) *soap.SoapHeader {
	return message.
		Header().
		To(url).
		ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").
		MaxEnvelopeSize(params.EnvelopeSize).
		Id(genUUID()).
		Locale(params.Locale).
		Timeout(params.Timeout)
}

//NewOpenShellRequest makes a new soap request
func NewOpenShellRequest(uri string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}

	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Create").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		AddOption(soap.NewHeaderOption("WINRS_NOPROFILE", "FALSE")).
		AddOption(soap.NewHeaderOption("WINRS_CODEPAGE", "65001")).
		Build()

	body := message.CreateBodyElement("Shell", soap.DOM_NS_WIN_SHELL)
	input := message.CreateElement(body, "InputStreams", soap.DOM_NS_WIN_SHELL)
	input.SetContent("stdin")
	output := message.CreateElement(body, "OutputStreams", soap.DOM_NS_WIN_SHELL)
	output.SetContent("stdout stderr")

	return message
}

// NewDeleteShellRequest ...
func NewDeleteShellRequest(uri, shellID string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Delete").
		ShellId(shellID).
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		Build()

	message.NewBody()

	return message
}

// NewExecuteCommandRequest exec command on specific shellID
func NewExecuteCommandRequest(uri, shellID, command string, arguments []string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		ShellId(shellID).
		AddOption(soap.NewHeaderOption("WINRS_CONSOLEMODE_STDIN", "TRUE")).
		AddOption(soap.NewHeaderOption("WINRS_SKIP_CMD_SHELL", "FALSE")).
		Build()

	body := message.CreateBodyElement("CommandLine", soap.DOM_NS_WIN_SHELL)

	// ensure special characters like & don't mangle the request XML
	command = "<![CDATA[" + command + "]]>"
	commandElement := message.CreateElement(body, "Command", soap.DOM_NS_WIN_SHELL)
	commandElement.SetContent(command)

	for _, arg := range arguments {
		arg = "<![CDATA[" + arg + "]]>"
		argumentsElement := message.CreateElement(body, "Arguments", soap.DOM_NS_WIN_SHELL)
		argumentsElement.SetContent(arg)
	}

	return message
}

//NewGetOutputRequest NewGetOutputRequest
func NewGetOutputRequest(uri, shellID, commandID, streams string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Receive").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		ShellId(shellID).
		Build()

	receive := message.CreateBodyElement("Receive", soap.DOM_NS_WIN_SHELL)
	desiredStreams := message.CreateElement(receive, "DesiredStream", soap.DOM_NS_WIN_SHELL)
	desiredStreams.SetAttr("CommandId", commandID)
	desiredStreams.SetContent(streams)

	return message
}

//NewSendInputRequest NewSendInputRequest
func NewSendInputRequest(uri, shellID, commandID string, input []byte, eof bool, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	message := soap.NewMessage()

	defaultHeaders(message, uri, params).
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Send").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		ShellId(shellID).
		Build()

	content := base64.StdEncoding.EncodeToString(input)

	send := message.CreateBodyElement("Send", soap.DOM_NS_WIN_SHELL)
	streams := message.CreateElement(send, "Stream", soap.DOM_NS_WIN_SHELL)
	streams.SetAttr("Name", "stdin")
	streams.SetAttr("CommandId", commandID)
	streams.SetContent(content)
	if eof {
		streams.SetAttr("End", "true")
	}
	return message
}

//NewSignalRequest NewSignalRequest
func NewSignalRequest(uri string, shellID string, commandID string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	message := soap.NewMessage()

	defaultHeaders(message, uri, params).
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Signal").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		ShellId(shellID).
		Build()

	signal := message.CreateBodyElement("Signal", soap.DOM_NS_WIN_SHELL)
	signal.SetAttr("CommandId", commandID)
	code := message.CreateElement(signal, "Code", soap.DOM_NS_WIN_SHELL)
	code.SetContent("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/signal/terminate")

	return message
}


// Open Powershell with MS-PSRP
func NewOpenPowerShellRequest(uri string, params *Parameters) (*soap.SoapMessage, string, string, string) {
	if params == nil {
		params = DefaultParameters
	}


	runspaceId := uuid.Must(uuid.NewV4()).String()
	pipelineId := uuid.Must(uuid.NewV4()).String()

	debugger.Println("In NewOpenPowerShellRe RunspaceID ->", runspaceId)
	debugger.Println("In NewOpenPowerShellRe PipelineId ->", pipelineId)

    creationXml, err := psrp.CreateShellInitData(runspaceId, pipelineId)
	if err != nil {
		debugger.Println("Error ------> ", err.Error())
		return nil, "", "", ""
	}

	messageId := genUUID()
	sessionId := genUUID()
	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Create").
		ResourceURI("http://schemas.microsoft.com/powershell/Microsoft.PowerShell").
		AddOption(soap.NewHeaderOption("protocolversion", "2.1")).
		ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").
		To("http://localhost/wsman").
		Id(messageId).
		SessionId(sessionId).
		Build()

	body := message.CreateBodyElement("Shell", soap.DOM_NS_WIN_SHELL)
	input := message.CreateElement(body, "InputStreams", soap.DOM_NS_WIN_SHELL)
	input.SetContent("stdin")
	output := message.CreateElement(body, "OutputStreams", soap.DOM_NS_WIN_SHELL)
	output.SetContent("stdout stderr")

	creationXmlElement := message.CreateElement(body, "creationXml", soap.DOM_NS_WIN_SHELL)
	creationXmlElement.SetContent(creationXml)

	return message, sessionId, runspaceId, pipelineId
}

func NewExecutePowerShellCommandRequest(uri, shellID, sessionId, runspaceId, pipelineId, command string, arguments []string, params *Parameters) *soap.SoapMessage {
	if params == nil {
		params = DefaultParameters
	}
	messageId := genUUID()
	message := soap.NewMessage()
	defaultHeaders(message, uri, params).
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command").
		ResourceURI("http://schemas.microsoft.com/powershell/Microsoft.PowerShell").
		AddOption(soap.NewHeaderOption("protocolversion", "2.1")).
		ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").
		To("http://localhost/wsman").
		ShellId(shellID).
		Id(messageId).
		SessionId(sessionId).
		Build()

	body := message.CreateBodyElement("CommandLine", soap.DOM_NS_WIN_SHELL)

	// Base64 Encoded Powershell XML

	debugger.Println("In NewExecuterPowerShellCommsn RunspaceID ->", runspaceId)
	debugger.Println("In NewExecuterPowerShellCommsn PipelineId ->", pipelineId)
	commandData, err := psrp.CreateCommandData(command, runspaceId, pipelineId) 
	if err != nil {
		debugger.Println("Error When Runnign Command ------> ", err.Error())
	}
	message.CreateElement(body, "Command", soap.DOM_NS_WIN_SHELL)
	ArgElement := message.CreateElement(body, "Arguments", soap.DOM_NS_WIN_SHELL)
	ArgElement.SetContent(commandData)


	return message
}
