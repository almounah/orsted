package socks

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/things-go/go-socks5"

	"orsted/beacon/utils"
)

var CONNECTED_SOCKS context.CancelFunc

type socks struct {
	Parentpeer utils.Peer
	ConnId     string
}

func (s *socks) Listen() (string, error) {
	var e utils.Envelope
	e.TypeEnv = "sockslisten"
	e.EnvId = utils.CurrentBeaconId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Listen", err.Error())
		return "", err
	}

	prepared, _ := s.Parentpeer.PrepareSocksData(tosend)
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

func (s *socks) Read(b []byte) (n int, err error) {
	var e utils.Envelope
	e.TypeEnv = "socksread"
	e.EnvId = s.ConnId
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

	prepared, _ := s.Parentpeer.PrepareSocksData(tosend)
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
		return 0, errors.New(string(envelopeReceived.EnvData))
	}

	copy(b, envelopeReceived.EnvData)
	if len(envelopeReceived.EnvData) != 0 {
		utils.Print("Very Succsefffully Read --> \n", envelopeReceived.EnvData)
	}

	return len(envelopeReceived.EnvData), nil
}

func (s *socks) Write(b []byte) (n int, err error) {
	utils.Print("In s5 Write")
	var e utils.Envelope
	e.TypeEnv = "sockswrite"
	e.EnvId = s.ConnId
	e.Chain = nil
	e.EnvData = b

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Read", err.Error())
		return 0, err
	}

	utils.Print("Will Write")
	prepared, _ := s.Parentpeer.PrepareSocksData(tosend)
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

func (s *socks) Close() error {
	utils.Print("In Close")
	var e utils.Envelope
	e.TypeEnv = "socksclose"
	e.EnvId = s.ConnId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in Close", err.Error())
		return err
	}

	prepared, _ := s.Parentpeer.PrepareSocksData(tosend)
	_, err = s.Parentpeer.SendRequest(prepared)

	return nil
}

func (c *socks) LocalAddr() net.Addr {
	utils.Print("In Local Addr")

	var e utils.Envelope
	e.TypeEnv = "sockslocaladdr"
	e.EnvId = c.ConnId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in LocalAdd", err.Error())
		return nil
	}

	prepared, _ := c.Parentpeer.PrepareSocksData(tosend)
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

func (c *socks) RemoteAddr() net.Addr {
	utils.Print("In Remote Addr")
	var e utils.Envelope
	e.TypeEnv = "socksremoteaddr"
	e.EnvId = c.ConnId
	e.Chain = nil
	e.EnvData = nil

	tosend, err := json.Marshal(&e)
	if err != nil {
		utils.Print("Error in marsheling in RemoteAddr", err.Error())
		return nil
	}

	prepared, _ := c.Parentpeer.PrepareSocksData(tosend)
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
func (c *socks) SetDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *socks) SetReadDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *socks) SetWriteDeadline(t time.Time) error {
	return nil
}

func StartSocksServer(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			utils.Print("Stopping SOCKS server loop")
			return
		default:
			// 1. Open a new WebSocket connection

			// 2. Start SOCKS server with the new WebSocket
			socksServer := socks5.NewServer()
			hangedSock := &socks{Parentpeer: utils.ParentPeer, ConnId: "Unset"}
			ConnId, err := hangedSock.Listen()
			if err != nil {
				utils.Print("Error Listening ", err.Error())
			}
			hangedSock.ConnId = ConnId
			go socksServer.ServeConn(hangedSock)

			// 3. Close the WebSocket after use
		}
	}
}

func StopSocksServer() {
}
