package mspsrp

import (
	"encoding/xml"
	"fmt"
)

// Define SOAP envelope structure
type Envelope struct {
	XMLName xml.Name `xml:"s:Envelope"`
	S       string   `xml:"xmlns:s,attr"`
	WSA     string   `xml:"xmlns:wsa,attr"`
	RSP     string   `xml:"xmlns:rsp,attr"`
	WSMAN   string   `xml:"xmlns:wsman,attr"`
	WSMV    string   `xml:"xmlns:wsmv,attr"`
	Header  Header   `xml:"s:Header"`
	Body    Body     `xml:"s:Body"`
}

type Header struct {
	ResourceURI      ResourceURI      `xml:"wsman:ResourceURI"`
	ReplyTo          ReplyTo          `xml:"wsa:ReplyTo"`
	To               string           `xml:"wsa:To"`
	Action           Action           `xml:"wsa:Action"`
	MessageID        string           `xml:"wsa:MessageID"`
	MaxEnvelopeSize  MaxEnvelopeSize  `xml:"wsman:MaxEnvelopeSize"`
	Locale           Locale           `xml:"wsman:Locale"`
	OperationTimeout string           `xml:"wsman:OperationTimeout"`
	OptionSet        OptionSet        `xml:"wsman:OptionSet"`
	DataLocale       DataLocale       `xml:"wsmv:DataLocale"`
	SessionID        SessionID        `xml:"wsmv:SessionId"`
	SelectorSet      *SelectorSet     `xml:"wsman:SelectorSet,omitempty"`
}

type ResourceURI struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type ReplyTo struct {
	Address Address `xml:"wsa:Address"`
}

type Address struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type Action struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type MaxEnvelopeSize struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type Locale struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Lang           string `xml:"xml:lang,attr"`
}

type OptionSet struct {
	MustUnderstand string   `xml:"s:mustUnderstand,attr"`
	Option         []Option `xml:"wsman:Option,omitempty"`
}

type Option struct {
	Name       string `xml:"Name,attr"`
	MustComply string `xml:"MustComply,attr,omitempty"`
	Value      string `xml:",chardata"`
}

type DataLocale struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Lang           string `xml:"xml:lang,attr"`
}

type SessionID struct {
	MustUnderstand string `xml:"s:mustUnderstand,attr"`
	Value          string `xml:",chardata"`
}

type SelectorSet struct {
	Selector []Selector `xml:"wsman:Selector"`
}

type Selector struct {
	Name  string `xml:"Name,attr"`
	Value string `xml:",chardata"`
}

type Body struct {
	Shell       *Shell       `xml:"rsp:Shell,omitempty"`
	Receive     *Receive     `xml:"rsp:Receive,omitempty"`
	CommandLine *CommandLine `xml:"rsp:CommandLine,omitempty"`
}

type Shell struct {
	ShellID       string `xml:"rsp:ShellId"`
	InputStreams  string `xml:"rsp:InputStreams"`
	OutputStreams string `xml:"rsp:OutputStreams"`
	CreationXML   string `xml:"creationXml"`
}

type Receive struct {
    DesiredStream DesiredStream `xml:"rsp:DesiredStream"`
}

type DesiredStream struct {
	CommandId string `xml:"CommandId,attr,omitempty"`
    Value     string `xml:",chardata"`
}

type CommandLine struct {
	Command   string `xml:"rsp:Command"`
	Arguments string `xml:"rsp:Arguments,omitempty"`
}

// Helper function to create base header
func createBaseHeader(action, messageID, operationTimeout, sessionID string) Header {
	return Header{
		ResourceURI: ResourceURI{
			MustUnderstand: "true",
			Value:          "http://schemas.microsoft.com/powershell/Microsoft.PowerShell",
		},
		ReplyTo: ReplyTo{
			Address: Address{
				MustUnderstand: "true",
				Value:          "http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous",
			},
		},
		To: "http://localhost/wsman",
		Action: Action{
			MustUnderstand: "true",
			Value:          action,
		},
		MessageID: messageID,
		MaxEnvelopeSize: MaxEnvelopeSize{
			MustUnderstand: "true",
			Value:          "153600",
		},
		Locale: Locale{
			MustUnderstand: "false",
			Lang:           "en-US",
		},
		OperationTimeout: operationTimeout,
		DataLocale: DataLocale{
			MustUnderstand: "false",
			Lang:           "en-US",
		},
		SessionID: SessionID{
			MustUnderstand: "false",
			Value:          sessionID,
		},
	}
}

// Create Shell request (first message)
func CreateShellRequest(creationXML string, messageID, sessionID string) *Envelope {
	header := createBaseHeader(
		"http://schemas.xmlsoap.org/ws/2004/09/transfer/Create",
		"uuid:"+messageID,
		"PT10S",
		"uuid:"+sessionID,
	)

	header.OptionSet = OptionSet{
		MustUnderstand: "true",
		Option: []Option{
			{
				Name:       "protocolversion",
				MustComply: "true",
				Value:      "2.1",
			},
		},
	}
	header.SelectorSet = &SelectorSet{Selector: []Selector{}}

	return &Envelope{
		S:     "http://www.w3.org/2003/05/soap-envelope",
		WSA:   "http://schemas.xmlsoap.org/ws/2004/08/addressing",
		RSP:   "http://schemas.microsoft.com/wbem/wsman/1/windows/shell",
		WSMAN: "http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd",
		WSMV:  "http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd",
		Header: header,
		Body: Body{
			Shell: &Shell{
				ShellID:       "http://localhost/wsman",
				InputStreams:  "stdin",
				OutputStreams: "stdout",
				CreationXML:   creationXML,
			},
		},
	}
}

// Create Command request (second message)
func CreateCommandRequest(commmandXML, shellID string, sessionId, messageId string) *Envelope {
	header := createBaseHeader(
		"http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command",
		"uuid:"+messageId,
		"PT1S",
		"uuid:"+sessionId,
	)

	header.OptionSet = OptionSet{
		MustUnderstand: "true",
	}
	header.SelectorSet = &SelectorSet{
		Selector: []Selector{
			{
				Name:  "ShellId",
				Value: shellID,
			},
		},
	}

	return &Envelope{
		S:     "http://www.w3.org/2003/05/soap-envelope",
		WSA:   "http://schemas.xmlsoap.org/ws/2004/08/addressing",
		RSP:   "http://schemas.microsoft.com/wbem/wsman/1/windows/shell",
		WSMAN: "http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd",
		WSMV:  "http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd",
		Header: header,
		Body: Body{
			CommandLine: &CommandLine{
				Command:   "",
				Arguments: commmandXML,
			},
		},
	}
}

// Create Receive request (third message)
// TODO Need to add COmmandID
func CreateReceiveRequest(commandId, shellID, messageId, sessionId string) *Envelope {
	header := createBaseHeader(
		"http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Receive",
		"uuid:"+messageId,
		"PT20S",
		"uuid:"+sessionId,
	)

	header.OptionSet = OptionSet{
		MustUnderstand: "true",
		Option: []Option{
			{
				Name:  "WSMAN_CMDSHELL_OPTION_KEEPALIVE",
				Value: "true",
			},
		},
	}
	header.SelectorSet = &SelectorSet{
		Selector: []Selector{
			{
				Name:  "ShellId",
				Value: shellID,
			},
		},
	}

	return &Envelope{
		S:     "http://www.w3.org/2003/05/soap-envelope",
		WSA:   "http://schemas.xmlsoap.org/ws/2004/08/addressing",
		RSP:   "http://schemas.microsoft.com/wbem/wsman/1/windows/shell",
		WSMAN: "http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd",
		WSMV:  "http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd",
		Header: header,
		Body: Body{
            Receive: &Receive{
                DesiredStream: DesiredStream{
                    CommandId: commandId,
                    Value:     "stdout",
                },
            },
        },
	}
}


// Envelope to String
func EnvelopeToString(envelope *Envelope) (string,error) {
	// Marshal to XML
	xmlData, err := xml.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling XML: %w", err)
	}

	// Add XML declaration
	soapRequest := []byte(xml.Header + string(xmlData))

	return string(soapRequest), nil
}

