package peers

import (
	"bufio"
	"net"

	"orsted/beacon/utils"
	"orsted/profiles"
)

// Peer that connect to IP Port via reverse connection
type TCPReversePeer struct {
	Id           string
	PeerType     string
	Conf         profiles.ProfileConfig
	Ip           string
	Port         string
	RealTimeConn net.Conn
}

// Http peer only require profile config
func NewTCPReversePeer(c profiles.ProfileConfig, ip string, port string) (*TCPReversePeer, error) {
	trp := &TCPReversePeer{}
	trp.PeerType = "tcp"
	trp.Conf = c
	trp.Ip = ip
	trp.Id = "-1"

	trp.Port = port

	return trp, nil
}

// Compare the Peer to current Beacon - Return Child or Parent or none
func (trp *TCPReversePeer) GetPeerLevel() string {
	for _, p := range utils.ChildPeers {
		if trp.Id == p.GetPeerID() {
			return "child"
		}
	}

	if utils.ParentPeer.GetPeerID() == trp.Id {
		return "parent"
	}

	return "none"
}

// Compare Peer connection to Current Beacon - Return TCP SMB or HTTP (for server)
func (trp *TCPReversePeer) GetPeerType() string {
	return trp.PeerType
}

// Get Peer (same as Beacon ID) (0 for Server, -1 for pending)
func (trp *TCPReversePeer) GetPeerID() string {
	return trp.Id
}

// Set Peer Id (same as Beacon ID) (used because Server decide Beacon Id)
func (trp *TCPReversePeer) SetPeerID(s string) string {
	trp.Id = s
	return s
}

func (trp *TCPReversePeer) GetPeerAddress() string {
	return trp.Ip + ":" + trp.Port
}

// Initialise TCP Dial COnn and return the underlying Conn
// Need to handler error and multiple TCP for pivot
func (trp *TCPReversePeer) GetRealTimeConn(beaconId string) (net.Conn, error) {
	utils.Print("Getting TCP RealTimeConn --> ")
	//if trp.RealTimeConn != nil {
	//	return trp.RealTimeConn, nil
	//}
	url := trp.GetPeerAddress()
	utils.Print("URL is --> ", url)

	netConn, err := net.Dial("tcp", url)
	if err != nil {
		utils.Print("Error in dialing --> ", err)
		return nil, err
	}


	// Do Handshake
	// 3 is for initialize
	t :=append([]byte{3}, []byte(beaconId)...) 
	netConn.Write(append(t, '\n'))

	buf := make([]byte, 2)
	netConn.Read(buf)
	utils.Print("Result of Handshake --> ", string(buf))

	trp.RealTimeConn = netConn

	return netConn, nil
}

// Send Data to Peer and get response
func (trp *TCPReversePeer) SendRequest(dataToSend []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", trp.Ip+":"+trp.Port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	dataToSend = append(dataToSend, byte('\n'))
	utils.Print("Will Hang TCP socket")
	_, err = conn.Write(dataToSend)

	utils.Print("TCP Socket hanged")
	//	resp, err := io.ReadAll(conn) // Read everything until connection closes
	resp, _ := bufio.NewReader(conn).ReadBytes(byte('\n'))
	utils.Print("You won't see me")
	if err != nil {
		panic(err)
	}

	//utils.Print(string(resp))

	return resp, nil
}

func (trp *TCPReversePeer) PrepareRegisterBeaconData(rawEnvelope []byte) ([]byte, error) {
	return append([]byte{1}, rawEnvelope...), nil
}

func (trp *TCPReversePeer) PrepareRetreiveTasksData(rawEnvelope []byte) ([]byte, error) {
	return append([]byte{1}, rawEnvelope...), nil
}

func (trp *TCPReversePeer) PrepareSendTaskResultsData(rawEnvelope []byte) ([]byte, error) {
	return append([]byte{1}, rawEnvelope...), nil
}

func (hp *TCPReversePeer) PrepareSocksData(rawEnvelope []byte) ([]byte, error) {
	return append([]byte{1}, rawEnvelope...), nil
}

func (hp *TCPReversePeer) PrepareAutorouteData(rawEnvelope []byte) ([]byte, error) {
	return append([]byte{1}, rawEnvelope...), nil
}

func (trp *TCPReversePeer) CleanReqFromPeerProtocol(data []byte) (rawEnvelope []byte, err error) {
	// TODO: Implement This when pivot include HTTP
	return data, nil
}

func (tcp *TCPReversePeer) CleanRespFromPeerProtocol(data []byte) (rawEnvelope []byte, err error) {
	return data, nil
}
