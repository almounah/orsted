package psrp

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// PSRP Message Types
const (
	SESSION_CAPABILITY   = 0x00010002
	INIT_RUNSPACEPOOL    = 0x00010004
	CREATE_PIPELINE      = 0x00021006
	PIPELINE_OUTPUT      = 0x00041004
	PIPELINE_STATE       = 0x00041006
	ERROR_RECORD         = 0x00041005
	PIPELINE_HOST_CALL   = 0x00041100
)

// PSObjectBuilder helps construct CLIXML objects

type PSRPMessage struct {
	MessageType uint32
	Data        []byte
}

// Create CLIXML elements
func psSimple(name, kind, value string) string {
	if value == "" {
		return fmt.Sprintf(`<Nil N="%s"/>`, name)
	}
	return fmt.Sprintf(`<%s N="%s">%s</%s>`, kind, name, value, kind)
}

func psEnum(name string, value int) string {
	return fmt.Sprintf(`<Obj N="%s"><I32>%d</I32></Obj>`, name, value)
}

func psStruct(name string, elements []string) string {
	attrs := ""
	if name != "" {
		attrs = fmt.Sprintf(` N="%s"`, name)
	}
	inner := strings.Join(elements, "")
	return fmt.Sprintf(`<Obj%s><MS>%s</MS></Obj>`, attrs, inner)
}

func psList(name string, elements []string) string {
	inner := strings.Join(elements, "")
	return fmt.Sprintf(`<Obj N="%s"><LST>%s</LST></Obj>`, name, inner)
}

// Create PowerShell capability object
func psCapability() string {
	return psStruct("", []string{
		psSimple("protocolversion", "Version", "2.1"),
		psSimple("PSVersion", "Version", "2.0"),
		psSimple("SerializationVersion", "Version", "1.1.0.10"),
	})
}

// Create runspace pool initialization object
func psRunspacePool() string {
	hostInfo := psStruct("HostInfo", []string{
		psSimple("_isHostNull", "B", "true"),
		psSimple("_isHostUINull", "B", "true"),
		psSimple("_isHostRawUINull", "B", "true"),
		psSimple("_useRunspaceHost", "B", "true"),
	})

	return psStruct("", []string{
		psSimple("MinRunspaces", "I32", "1"),
		psSimple("MaxRunspaces", "I32", "1"),
		psEnum("PSThreadOptions", 0),
		psEnum("ApartmentState", 2),
		hostInfo,
		psSimple("ApplicationArguments", "Nil", ""),
	})
}

// Create command object
func psCommand(cmd string, args map[string]string) string {
	argsList := []string{}
	for k, v := range args {
		argObj := psStruct("", []string{
			psSimple("N", "S", k),
			psSimple("V", "S", v),
		})
		argsList = append(argsList, argObj)
	}

	return psStruct("", []string{
		psSimple("Cmd", "S", cmd),
		psList("Args", argsList),
		psSimple("IsScript", "B", "false"),
		psSimple("UseLocalScope", "Nil", ""),
		psEnum("MergeMyResult", 0),
		psEnum("MergeToResult", 0),
		psEnum("MergePreviousResults", 0),
		psEnum("MergeError", 0),
		psEnum("MergeWarning", 0),
		psEnum("MergeVerbose", 0),
		psEnum("MergeDebug", 0),
		psEnum("MergeInformation", 0),
	})
}

// Create pipeline with commands
func psCreatePipeline(commands []string) string {
	hostInfo := psStruct("HostInfo", []string{
		psSimple("_isHostNull", "B", "true"),
		psSimple("_isHostUINull", "B", "true"),
		psSimple("_isHostRawUINull", "B", "true"),
		psSimple("_useRunspaceHost", "B", "true"),
	})

	powerShell := psStruct("PowerShell", []string{
		psSimple("IsNested", "B", "false"),
		psSimple("RedirectShellErrorOutputPipe", "B", "false"),
		psSimple("ExtraCmds", "Nil", ""),
		psSimple("History", "Nil", ""),
		psList("Cmds", commands),
	})

	return psStruct("", []string{
		psSimple("NoInput", "B", "true"),
		psSimple("AddToHistory", "B", "false"),
		psSimple("IsNested", "B", "false"),
		psEnum("ApartmentState", 2),
		psEnum("RemoteStreamOptions", 15),
		hostInfo,
		powerShell,
	})
}

// Fragment PSRP messages into binary format
func fragmentPSRP(messages []PSRPMessage, runspaceID, pipelineID string) ([]byte, error) {
	fragments := []byte{}
	objectID := uint64(1)

	runspaceUUID, err := uuid.Parse(runspaceID)
	if err != nil {
		return nil, err
	}

	pipelineUUID, err := uuid.Parse(pipelineID)
	if err != nil {
		return nil, err
	}

	for _, msg := range messages {
		// Message data structure
		msgData := make([]byte, 4)
		binary.LittleEndian.PutUint32(msgData, 0x00002) // Destination

		msgType := make([]byte, 4)
		binary.LittleEndian.PutUint32(msgType, msg.MessageType)
		msgData = append(msgData, msgType...)

		// Add UUIDs (little-endian format)
		runspaceBytes, _ := runspaceUUID.MarshalBinary()
		pipelineBytes, _ := pipelineUUID.MarshalBinary()
		msgData = append(msgData, runspaceBytes...)
		msgData = append(msgData, pipelineBytes...)

		// Add CLIXML payload
		msgData = append(msgData, msg.Data...)

		// Fragment header
		fragment := make([]byte, 8)
		binary.BigEndian.PutUint64(fragment, objectID)

		fragmentID := make([]byte, 8)
		binary.BigEndian.PutUint64(fragmentID, 0) // Fragment ID (0 for single fragment)
		fragment = append(fragment, fragmentID...)

		fragment = append(fragment, 0x03) // Start and End flag

		msgLen := make([]byte, 4)
		binary.BigEndian.PutUint32(msgLen, uint32(len(msgData)))
		fragment = append(fragment, msgLen...)

		fragment = append(fragment, msgData...)
		fragments = append(fragments, fragment...)
		objectID++
	}

	return fragments, nil
}

// Create shell initialization data (for creationXml)
func CreateShellInitData(runspaceID, pipelineID string) (string, error) {
	messages := []PSRPMessage{
		{
			MessageType: SESSION_CAPABILITY,
			Data:        []byte(psCapability()),
		},
		{
			MessageType: INIT_RUNSPACEPOOL,
			Data:        []byte(psRunspacePool()),
		},
	}

	data, err := fragmentPSRP(messages, runspaceID, pipelineID)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

// Create command execution data (for Command arguments)
func CreateCommandData(command string, runspaceID, pipelineID string) (string, error) {
	// Create commands for pipeline
	cmdObj := psCommand(command, map[string]string{})
	outStringCmd := psCommand("Out-String", map[string]string{
		"Stream": "",
	})

	pipeline := psCreatePipeline([]string{cmdObj, outStringCmd})

	messages := []PSRPMessage{
		{
			MessageType: CREATE_PIPELINE,
			Data:        []byte(pipeline),
		},
	}

	data, err := fragmentPSRP(messages, runspaceID, pipelineID)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

