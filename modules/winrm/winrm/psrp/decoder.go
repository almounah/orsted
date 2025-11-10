package psrp

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"strings"
)

// PSRPFragment represents a single PSRP fragment
type PSRPFragment struct {
	ObjectID   uint64
	FragmentID uint64
	StartEnd   byte
	MessageLen uint32
	Destination uint32
	MessageType uint32
	RunspaceID [16]byte
	PipelineID [16]byte
	Payload    []byte
}

// PSRPOutput represents parsed output from PSRP
type PSRPOutput struct {
	Stdout   string
	Stderr   string
	Error    string
	Warning  string
	Verbose  string
	Info     string
	State    string
	ExitCode int
}

// ParsePSRPResponse decodes base64-encoded PSRP fragments
func ParsePSRPResponse(base64Data string) ([]PSRPFragment, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode error: %v", err)
	}

	fragments := []PSRPFragment{}
	offset := 0

	for offset < len(data) {
		if len(data[offset:]) < 21 {
			break
		}

		fragment := PSRPFragment{}
		fragment.ObjectID = binary.BigEndian.Uint64(data[offset : offset+8])
		fragment.FragmentID = binary.BigEndian.Uint64(data[offset+8 : offset+16])
		fragment.StartEnd = data[offset+16]
		fragment.MessageLen = binary.BigEndian.Uint32(data[offset+17 : offset+21])

		offset += 21

		if len(data[offset:]) < int(fragment.MessageLen) {
			return nil, fmt.Errorf("incomplete fragment at offset %d", offset)
		}

		messageData := data[offset : offset+int(fragment.MessageLen)]
		offset += int(fragment.MessageLen)

		if len(messageData) < 40 {
			continue
		}

		fragment.Destination = binary.LittleEndian.Uint32(messageData[0:4])
		fragment.MessageType = binary.LittleEndian.Uint32(messageData[4:8])
		copy(fragment.RunspaceID[:], messageData[8:24])
		copy(fragment.PipelineID[:], messageData[24:40])
		fragment.Payload = messageData[40:]

		fragments = append(fragments, fragment)
	}

	return fragments, nil
}

// ExtractTextFromCLIXML extracts text content from CLIXML
func ExtractTextFromCLIXML(clixml []byte) string {
	decoder := xml.NewDecoder(strings.NewReader(string(clixml)))
	var result strings.Builder

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.CharData:
			text := strings.TrimSpace(string(t))
			if text != "" {
				result.WriteString(text)
			}
		}
	}

	return result.String()
}

// ParsePSRPOutput processes PSRP fragments and extracts output
func ParsePSRPOutput(base64Data string) (*PSRPOutput, error) {
	fragments, err := ParsePSRPResponse(base64Data)
	if err != nil {
		return nil, err
	}

	output := &PSRPOutput{}

	for _, frag := range fragments {
		switch frag.MessageType {
		case PIPELINE_OUTPUT: // 0x00041004 - stdout
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Stdout += text

		case ERROR_RECORD: // 0x00041005 - errors
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Error += text

		case PIPELINE_STATE: // 0x00041006 - pipeline state
			text := ExtractTextFromCLIXML(frag.Payload)
			output.State = text

		case 0x00041007: // DEBUG_RECORD
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Verbose += text

		case 0x00041008: // VERBOSE_RECORD
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Verbose += text

		case 0x00041009: // WARNING_RECORD
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Warning += text

		case 0x00041011: // INFORMATION_RECORD (Write-Host)
			text := ExtractTextFromCLIXML(frag.Payload)
			output.Info += text
		}
	}

	return output, nil
}

