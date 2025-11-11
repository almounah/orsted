package pwshxml

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// MS-PSRP message type constants
const (
	SESSION_CAPABILITY = 0x00010002
	INIT_RUNSPACEPOOL  = 0x00010004
	CREATE_PIPELINE    = 0x00021006
)

// CreateCreationXML generates the base64-encoded creationXml for WinRM shell creation
// runspaceID and pipelineID should be valid UUID strings (e.g., "550e8400-e29b-41d4-a716-446655440000")
func CreateCreationXML(runspaceID, pipelineID string) (string, error) {
	runspaceUUID, err := uuid.Parse(runspaceID)
	if err != nil {
		return "", fmt.Errorf("invalid runspace ID: %w", err)
	}

	pipelineUUID, err := uuid.Parse(pipelineID)
	if err != nil {
		return "", fmt.Errorf("invalid pipeline ID: %w", err)
	}

	// Build the two required messages
	capabilityXML := buildCapabilityXML()
	runspacePoolXML := buildRunspacePoolXML()

	// Fragment both messages
	var fragments bytes.Buffer
	objectID := uint64(1)

	if err := writeFragment(&fragments, SESSION_CAPABILITY, capabilityXML, runspaceUUID, pipelineUUID, objectID); err != nil {
		return "", err
	}
	objectID++

	if err := writeFragment(&fragments, INIT_RUNSPACEPOOL, runspacePoolXML, runspaceUUID, pipelineUUID, objectID); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(fragments.Bytes()), nil
}

// uuidToLittleEndian converts a UUID to little-endian byte format (Windows GUID format)
// This matches Python's uuid.UUID().bytes_le behavior
func uuidToLittleEndian(u uuid.UUID) []byte {
	b := make([]byte, 16)
	// First 4 bytes (time_low) - reverse
	b[0], b[1], b[2], b[3] = u[3], u[2], u[1], u[0]
	// Next 2 bytes (time_mid) - reverse
	b[4], b[5] = u[5], u[4]
	// Next 2 bytes (time_hi_version) - reverse
	b[6], b[7] = u[7], u[6]
	// Last 8 bytes (clock_seq and node) - keep as is
	copy(b[8:], u[8:])
	return b
}

// writeFragment writes a single MS-PSRP message fragment
func writeFragment(buf *bytes.Buffer, msgType uint32, xmlData string, runspaceID, pipelineID uuid.UUID, objectID uint64) error {
	// Build message data: destination + msgType + runspaceID + pipelineID + XML
	msgBuf := new(bytes.Buffer)

	// Destination (0x00002) - little endian
	if err := binary.Write(msgBuf, binary.LittleEndian, uint32(0x00002)); err != nil {
		return err
	}

	// Message type - little endian
	if err := binary.Write(msgBuf, binary.LittleEndian, msgType); err != nil {
		return err
	}

	// Runspace UUID - little-endian format (Windows GUID format)
	// Python uses: uuid.UUID(runspace_id).bytes_le
	runspaceBytes := uuidToLittleEndian(runspaceID)
	msgBuf.Write(runspaceBytes)

	// Pipeline UUID - little-endian format (Windows GUID format)
	pipelineBytes := uuidToLittleEndian(pipelineID)
	msgBuf.Write(pipelineBytes)

	// XML payload
	msgBuf.WriteString(xmlData)

	msgData := msgBuf.Bytes()

	// Write fragment header
	// Object ID - big endian uint64
	if err := binary.Write(buf, binary.BigEndian, objectID); err != nil {
		return err
	}

	// Fragment ID - big endian uint64 (0 for complete message)
	if err := binary.Write(buf, binary.BigEndian, uint64(0)); err != nil {
		return err
	}

	// Start/End flags - byte (3 = start and end, complete message)
	buf.WriteByte(3)

	// Message length - big endian uint32
	if err := binary.Write(buf, binary.BigEndian, uint32(len(msgData))); err != nil {
		return err
	}

	// Message data
	buf.Write(msgData)

	return nil
}

// buildCapabilityXML builds the SESSION_CAPABILITY XML
func buildCapabilityXML() string {
	var b strings.Builder
	b.WriteString("<Obj>")
	b.WriteString("<MS>")
	b.WriteString(`<Version N="protocolversion">2.1</Version>`)
	b.WriteString(`<Version N="PSVersion">2.0</Version>`)
	b.WriteString(`<Version N="SerializationVersion">1.1.0.10</Version>`)
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	return b.String()
}

// buildRunspacePoolXML builds the INIT_RUNSPACEPOOL XML
func buildRunspacePoolXML() string {
	var b strings.Builder
	b.WriteString("<Obj>")
	b.WriteString("<MS>")
	b.WriteString(`<I32 N="MinRunspaces">1</I32>`)
	b.WriteString(`<I32 N="MaxRunspaces">1</I32>`)
	b.WriteString(`<Obj N="PSThreadOptions"><I32>0</I32></Obj>`)
	b.WriteString(`<Obj N="ApartmentState"><I32>2</I32></Obj>`)
	b.WriteString(`<Obj N="HostInfo">`)
	b.WriteString("<MS>")
	b.WriteString(`<B N="_isHostNull">true</B>`)
	b.WriteString(`<B N="_isHostUINull">true</B>`)
	b.WriteString(`<B N="_isHostRawUINull">true</B>`)
	b.WriteString(`<B N="_useRunspaceHost">true</B>`)
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	b.WriteString(`<Nil N="ApplicationArguments"/>`)
	b.WriteString("</MS>")
	b.WriteString("</Obj>")
	return b.String()
}

// --- Optional: XML-based approach (more flexible but slower) ---

type xmlElement struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Inner   []xmlElement
}

func (e xmlElement) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	start.Name = e.XMLName
	start.Attr = e.Attrs
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	if e.Content != "" {
		if err := enc.EncodeToken(xml.CharData(e.Content)); err != nil {
			return err
		}
	}
	for _, child := range e.Inner {
		if err := enc.Encode(child); err != nil {
			return err
		}
	}
	return enc.EncodeToken(xml.EndElement{Name: start.Name})
}

