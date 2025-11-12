package pwshxml

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"strings"
)

// PSRP Message Type Constants
const (
	PUBLIC_KEY                 = 0x00010005
	ENCRYPTED_SESSION_KEY      = 0x00010006
	PUBLIC_KEY_REQUEST         = 0x00010007
	CONNECT_RUNSPACEPOOL       = 0x00010008
	RUNSPACEPOOL_INIT_DATA     = 0x0002100B
	RESET_RUNSPACE_STATE       = 0x0002100C
	SET_MAX_RUNSPACES          = 0x00021002
	SET_MIN_RUNSPACES          = 0x00021003
	RUNSPACE_AVAILABILITY      = 0x00021004
	RUNSPACEPOOL_STATE         = 0x00021005
	GET_AVAILABLE_RUNSPACES    = 0x00021007
	USER_EVENT                 = 0x00021008
	APPLICATION_PRIVATE_DATA   = 0x00021009
	GET_COMMAND_METADATA       = 0x0002100A
	RUNSPACEPOOL_HOST_CALL     = 0x00021100
	RUNSPACEPOOL_HOST_RESPONSE = 0x00021101
	PIPELINE_INPUT             = 0x00041002
	END_OF_PIPELINE_INPUT      = 0x00041003
	PIPELINE_OUTPUT            = 0x00041004
	ERROR_RECORD               = 0x00041005
	PIPELINE_STATE             = 0x00041006
	DEBUG_RECORD               = 0x00041007
	VERBOSE_RECORD             = 0x00041008
	WARNING_RECORD             = 0x00041009
	PROGRESS_RECORD            = 0x00041010
	INFORMATION_RECORD         = 0x00041011
	PIPELINE_HOST_CALL         = 0x00041100
	PIPELINE_HOST_RESPONSE     = 0x00041101
)

// PSRPMessage represents a parsed PSRP message
type PSRPMessage struct {
	Type    uint32
	XMLData string
}

// PSRPDecoder handles defragmentation and parsing of PSRP messages
type PSRPDecoder struct {
	fragmentBuffer map[uint64][]byte
}

// NewPSRPDecoder creates a new PSRP decoder
func NewPSRPDecoder() *PSRPDecoder {
	return &PSRPDecoder{
		fragmentBuffer: make(map[uint64][]byte),
	}
}

// Defragment parses a stream of fragmented PSRP messages
func (d *PSRPDecoder) Defragment(stream []byte) ([]PSRPMessage, error) {
	var messages []PSRPMessage
	var fragments [][]byte

	buf := stream
	for len(buf) >= 21 {
		// Parse fragment header (21 bytes)
		objectID := binary.BigEndian.Uint64(buf[0:8])
		fragmentID := binary.BigEndian.Uint64(buf[8:16])
		startEnd := buf[16]
		msgLen := binary.BigEndian.Uint32(buf[17:21])

		if len(buf) < 21+int(msgLen) {
			break
		}

		partial := buf[21 : 21+msgLen]
		buf = buf[21+msgLen:]

		// Handle fragmentation
		switch startEnd {
		case 3: // Start and end (complete message)
			fragments = append(fragments, partial)
		case 1: // Start
			d.fragmentBuffer[objectID] = partial
		case 0: // Middle
			if existing, ok := d.fragmentBuffer[objectID]; ok {
				d.fragmentBuffer[objectID] = append(existing, partial...)
			}
		case 2: // End
			if existing, ok := d.fragmentBuffer[objectID]; ok {
				fragments = append(fragments, append(existing, partial...))
				delete(d.fragmentBuffer, objectID)
			}
		}

		_ = fragmentID // Unused for now
	}

	// Parse complete fragments
	for _, frag := range fragments {
		if len(frag) < 40 {
			continue
		}

		// Parse message header (first 8 bytes)
		// destination := binary.LittleEndian.Uint32(frag[0:4])
		msgType := binary.LittleEndian.Uint32(frag[4:8])

		// Skip UUIDs (32 bytes: runspace + pipeline)
		// XML data starts at offset 40
		if len(frag) > 40 {
			messages = append(messages, PSRPMessage{
				Type:    msgType,
				XMLData: string(frag[40:]),
			})
		}
	}

	return messages, nil
}

// ExtractOutput extracts readable output from PSRP messages
func ExtractOutput(messages []PSRPMessage) []string {
	var output []string

	for _, msg := range messages {
		switch msg.Type {
		case PIPELINE_OUTPUT:
			// Parse XML and extract text
			if text := extractTextFromXML(msg.XMLData); text != "" {
				output = append(output, text)
			}

		case ERROR_RECORD:
			// Extract error message
			if text := extractXMLValue(msg.XMLData, "ToString"); text != "" {
				output = append(output, "[ERROR] "+text)
			}

		case WARNING_RECORD:
			// Extract warning message
			if text := extractXMLValue(msg.XMLData, "ToString"); text != "" {
				output = append(output, "[WARNING] "+text)
			}

		case VERBOSE_RECORD:
			// Extract verbose message
			if text := extractXMLValue(msg.XMLData, "ToString"); text != "" {
				output = append(output, "[VERBOSE] "+text)
			}

		case INFORMATION_RECORD:
			// Extract information message (Write-Host)
			if text := extractXMLValue(msg.XMLData, "Message"); text != "" {
				output = append(output, text)
			}

		case PIPELINE_STATE:
			// Check if pipeline failed (state=5) and extract error
			if strings.Contains(msg.XMLData, "<I32 N=\"PipelineState\">5</I32>") {
				// State 5 = Failed, extract exception message
				if text := extractFirstToString(msg.XMLData); text != "" {
					output = append(output, "[ERROR] "+text)
				}
			}
		}
	}

	return output
}

// extractTextFromXML extracts text content from CLIXML
func extractTextFromXML(xmlData string) string {
	// Simple text extraction - parse <S> elements
	type S struct {
		Text string `xml:",chardata"`
	}
	
	var s S
	if err := xml.Unmarshal([]byte(xmlData), &s); err == nil {
		return decodeUTF(s.Text)
	}
	
	return ""
}

// extractXMLValue extracts a specific value from nested XML
func extractXMLValue(xmlData, elementName string) string {
	// Simple extraction using string search
	// Look for <S N="elementName">value</S> pattern
	startTag := fmt.Sprintf(`<S N="%s">`, elementName)
	endTag := "</S>"
	
	startIdx := strings.Index(xmlData, startTag)
	if startIdx == -1 {
		// Try without N attribute
		startTag = "<ToString>"
		endTag = "</ToString>"
		startIdx = strings.Index(xmlData, startTag)
		if startIdx == -1 {
			return ""
		}
	}
	
	startIdx += len(startTag)
	endIdx := strings.Index(xmlData[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}
	
	return decodeUTF(xmlData[startIdx : startIdx+endIdx])
}

// extractFirstToString finds the first <ToString> element in XML
func extractFirstToString(xmlData string) string {
	startTag := "<ToString>"
	endTag := "</ToString>"
	
	startIdx := strings.Index(xmlData, startTag)
	if startIdx == -1 {
		return ""
	}
	
	startIdx += len(startTag)
	endIdx := strings.Index(xmlData[startIdx:], endTag)
	if endIdx == -1 {
		return ""
	}
	
	return decodeUTF(xmlData[startIdx : startIdx+endIdx])
}

// decodeUTF decodes _xHHHH_ encoded strings in CLIXML
// e.g., "_x000A_" represents newline
func decodeUTF(s string) string {
	// Simple implementation - just handle common cases
	s = strings.ReplaceAll(s, "_x000A_", "\n")
	s = strings.ReplaceAll(s, "_x000D_", "\r")
	s = strings.ReplaceAll(s, "_x0009_", "\t")
	return s
}

// GetMessageTypeName returns the human-readable name for a message type
func GetMessageTypeName(msgType uint32) string {
	names := map[uint32]string{
		SESSION_CAPABILITY:         "SESSION_CAPABILITY",
		INIT_RUNSPACEPOOL:          "INIT_RUNSPACEPOOL",
		PUBLIC_KEY:                 "PUBLIC_KEY",
		ENCRYPTED_SESSION_KEY:      "ENCRYPTED_SESSION_KEY",
		PUBLIC_KEY_REQUEST:         "PUBLIC_KEY_REQUEST",
		CONNECT_RUNSPACEPOOL:       "CONNECT_RUNSPACEPOOL",
		RUNSPACEPOOL_INIT_DATA:     "RUNSPACEPOOL_INIT_DATA",
		RESET_RUNSPACE_STATE:       "RESET_RUNSPACE_STATE",
		SET_MAX_RUNSPACES:          "SET_MAX_RUNSPACES",
		SET_MIN_RUNSPACES:          "SET_MIN_RUNSPACES",
		RUNSPACE_AVAILABILITY:      "RUNSPACE_AVAILABILITY",
		RUNSPACEPOOL_STATE:         "RUNSPACEPOOL_STATE",
		CREATE_PIPELINE:            "CREATE_PIPELINE",
		GET_AVAILABLE_RUNSPACES:    "GET_AVAILABLE_RUNSPACES",
		USER_EVENT:                 "USER_EVENT",
		APPLICATION_PRIVATE_DATA:   "APPLICATION_PRIVATE_DATA",
		GET_COMMAND_METADATA:       "GET_COMMAND_METADATA",
		RUNSPACEPOOL_HOST_CALL:     "RUNSPACEPOOL_HOST_CALL",
		RUNSPACEPOOL_HOST_RESPONSE: "RUNSPACEPOOL_HOST_RESPONSE",
		PIPELINE_INPUT:             "PIPELINE_INPUT",
		END_OF_PIPELINE_INPUT:      "END_OF_PIPELINE_INPUT",
		PIPELINE_OUTPUT:            "PIPELINE_OUTPUT",
		ERROR_RECORD:               "ERROR_RECORD",
		PIPELINE_STATE:             "PIPELINE_STATE",
		DEBUG_RECORD:               "DEBUG_RECORD",
		VERBOSE_RECORD:             "VERBOSE_RECORD",
		WARNING_RECORD:             "WARNING_RECORD",
		PROGRESS_RECORD:            "PROGRESS_RECORD",
		INFORMATION_RECORD:         "INFORMATION_RECORD",
		PIPELINE_HOST_CALL:         "PIPELINE_HOST_CALL",
		PIPELINE_HOST_RESPONSE:     "PIPELINE_HOST_RESPONSE",
	}
	
	if name, ok := names[msgType]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(0x%08X)", msgType)
}

// DecodeStreams is a helper function that takes base64-encoded PSRP streams
// (from <rsp:Stream> elements) and returns all readable output
func DecodeStreams(base64Streams []string) ([]string, error) {
	decoder := NewPSRPDecoder()
	var allOutput []string

	for _, encodedStream := range base64Streams {
		// Base64 decode
		decodedStream, err := base64.StdEncoding.DecodeString(encodedStream)
		if err != nil {
			continue // Skip invalid base64
		}

		// Defragment PSRP messages
		messages, err := decoder.Defragment(decodedStream)
		if err != nil {
			continue
		}

		// Extract readable output
		output := ExtractOutput(messages)
		allOutput = append(allOutput, output...)
	}

	return allOutput, nil
}

