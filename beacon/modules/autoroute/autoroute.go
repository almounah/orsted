package autoroute

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/hashicorp/yamux"

	"orsted/beacon/modules/autoroute/agent"
	"orsted/beacon/utils"
)

var CONNECTED_SOCKS context.CancelFunc

type autorouteConn struct {
	Parentpeer utils.Peer
	ConnId     string
}

func (s *autorouteConn) Listen() (string, error) {
	var e utils.Envelope
	e.TypeEnv = "autoroutelisten"
	e.EnvId = utils.CurrentBeaconId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Listen", err.Error())
		return "", err
	}

	prepared, _ := s.Parentpeer.PrepareAutorouteData(tosend)
	resp, err := s.Parentpeer.SendRequest(prepared)

	raw, _ := s.Parentpeer.CleanRespFromPeerProtocol(resp)

	if raw != nil && len(raw) > 0 {
		utils.Print("Raw Receive in Listen ", string(raw))
	}
	var envelopeReceived utils.Envelope
	err = json.Unmarshal(raw, &envelopeReceived)
	if err != nil {
		utils.Print("Error in Unmarsheling in Listen", err.Error())
		return "", err
	}

	// TODO: Change Id and put it in body
	return envelopeReceived.EnvId, nil
}

func (s *autorouteConn) Read(b []byte) (n int, err error) {
	var e utils.Envelope
	e.TypeEnv = "autorouteread"
	e.EnvId = utils.CurrentBeaconId
	e.Chain = nil
	bytes := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(bytes, int64(len(b)))
	e.EnvData = bytes
	utils.Print("Connection Id -> ", e.EnvId)

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Read", err.Error())
		return 0, err
	}

	prepared, _ := s.Parentpeer.PrepareAutorouteData(tosend)
	utils.Print("Connection Id", s.ConnId)
	utils.Print("Will send \n", string(prepared))
	resp, err := s.Parentpeer.SendRequest(prepared)

	raw, _ := s.Parentpeer.CleanRespFromPeerProtocol(resp)

	if raw != nil && len(raw) > 0 {
		// utils.Print("Raw Receive in Read ", string(raw))
	}
	var envelopeReceived utils.Envelope
	err = json.Unmarshal(raw, &envelopeReceived)
	if err != nil {
		utils.Print("Error in Unmarsheling in Read", err.Error())
		return 0, err
	}

	if envelopeReceived.TypeEnv == "sockserror" {
		return 0, io.EOF
		//return 0, errors.New(string(envelopeReceived.EnvData))
	}

	copy(b, envelopeReceived.EnvData)
	if len(envelopeReceived.EnvData) != 0 {
		utils.Print("Very Succsefffully Read --> \n", envelopeReceived.EnvData)
	}

	return len(envelopeReceived.EnvData), nil
}

func (s *autorouteConn) Write(b []byte) (n int, err error) {
	utils.Print("In s5 Write")
	var e utils.Envelope
	e.TypeEnv = "autoroutewrite"
	e.EnvId = utils.CurrentBeaconId
	e.Chain = nil
	e.EnvData = b

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Read", err.Error())
		return 0, err
	}

	utils.Print("Will Write")
	prepared, _ := s.Parentpeer.PrepareAutorouteData(tosend)
	resp, err := s.Parentpeer.SendRequest(prepared)

	utils.Print("Received in Write \n ", string(resp))

	raw, _ := s.Parentpeer.CleanRespFromPeerProtocol(resp)
	utils.Print("Cleaned from Peer Protocol in Write \n ", string(resp))
	// TODO: Add n reteival over network
	var envelopeReceived utils.Envelope
	err = json.Unmarshal(raw, &envelopeReceived)
	if err != nil {
		utils.Print("Error in Unmarsheling in Write", err.Error())
		return 0, err
	}
	if envelopeReceived.TypeEnv == "sockserror" {
		return 0, errors.New(string(envelopeReceived.EnvData))
	}

	return len(b), nil
}

func (s *autorouteConn) Close() error {
	utils.Print("In Close")
	var e utils.Envelope
	e.TypeEnv = "autorouteclose"
	e.EnvId = utils.CurrentBeaconId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Close", err.Error())
		return err
	}

	prepared, _ := s.Parentpeer.PrepareAutorouteData(tosend)
	_, err = s.Parentpeer.SendRequest(prepared)

	return nil
}

func (c *autorouteConn) LocalAddr() net.Addr {
	utils.Print("In Local Addr")

	var e utils.Envelope
	e.TypeEnv = "autoroutelocaladdr"
	e.EnvId = c.ConnId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in LocalAdd", err.Error())
		return nil
	}

	prepared, _ := c.Parentpeer.PrepareAutorouteData(tosend)
	resp, err := c.Parentpeer.SendRequest(prepared)

	var envelopeReceived utils.Envelope
	raw, _ := c.Parentpeer.CleanRespFromPeerProtocol(resp)
	err = json.Unmarshal(raw, &envelopeReceived)
	if err != nil {
		utils.Print("Error in Unmarsheling in LocalAddr", err.Error())
		return nil
	}

	addressString := string(envelopeReceived.EnvData)
	utils.Print("Recived this custom address stringg", addressString)

	// TODO: Fix this to handle TCP and other just in case
	r := strings.SplitN(addressString, "//", 2)
	address := r[1]

	newNet, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil
	}

	return newNet
}

func (c *autorouteConn) RemoteAddr() net.Addr {
	utils.Print("In Remote Addr")
	var e utils.Envelope
	e.TypeEnv = "autorouteremoteaddr"
	e.EnvId = c.ConnId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in RemoteAddr", err.Error())
		return nil
	}

	prepared, _ := c.Parentpeer.PrepareAutorouteData(tosend)
	resp, err := c.Parentpeer.SendRequest(prepared)

	var envelopeReceived utils.Envelope
	raw, _ := c.Parentpeer.CleanRespFromPeerProtocol(resp)
	err = json.Unmarshal(raw, &envelopeReceived)
	if err != nil {
		utils.Print("Error in Unmarsheling in RemoteAddr", err.Error())
		return nil
	}

	addressString := string(envelopeReceived.EnvData)
	utils.Print("Recived this custom address string", addressString)

	// TODO: Fix this to handle TCP and other just in case
	r := strings.SplitN(addressString, "//", 2)
	address := r[1]

	newNet, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil
	}

	return newNet
}

// TODO impl
func (c *autorouteConn) SetDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *autorouteConn) SetReadDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *autorouteConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func StartAutorouting(ctx context.Context) error {
	autorouteConnVar := autorouteConn{Parentpeer: utils.ParentPeer, ConnId: utils.CurrentBeaconId}
	fmt.Println("Sleeping for 5 seconds")
	fmt.Println("This is updated beacon")
	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 120*time.Second // ping every 2min
	cfg.ConnectionWriteTimeout = 120*time.Second
	yamuxConn, err := yamux.Server(&autorouteConnVar, cfg)
	if err != nil {
		fmt.Println("Error Creating Server ", err)
		return err
	}
	fmt.Println("Started Yamux Server Successfully")

	for {
	conn, err := yamuxConn.Accept()
	if err != nil {
		fmt.Println("Error Accepting Yamux Conn", err)
		return err
	}
	utils.Print("Accepted Yamux Session successfully, will start handling")
	go agent.HandleConn(conn)
	}
	return nil

}

func StopSocksServer() {
}
