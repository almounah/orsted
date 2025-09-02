package socks

import (
	"encoding/binary"
	"net"
	"sync"

	"github.com/google/uuid"

	"orsted/protobuf/orstedrpc"
	"orsted/server/utils"
)

type SOCKS struct {
	Ip         string
	Port       string
	Mu         sync.Mutex       // Changed RWMutex to Mutex for simplicity
	ConnMap    map[string]net.Conn
	ToSendConn map[string]net.Conn
	Cond       *sync.Cond
}

var SOCKS_LISTENER SOCKS

// StartNewSocksServer listens for TCP connections on the given IP and port
func StartNewSocksServer(beaconId string, ip string, port string) error {
	// Suppose we have an HTTP listener; otherwise return error.
	//if len(listeners.LISTENERS_HTTP_LIST) == 0 { // Make sure LISTENERS_HTTP_LIST is defined properly elsewhere
	//	return errors.New("No HTTP Listener present")
	//}

    utils.PrintInfo("Starting Socks on ", ip, port)
	l, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		utils.PrintDebug("Error Listining", err.Error())
		return err
	}

	// Initialize listener settings
	SOCKS_LISTENER.Ip = ip
	SOCKS_LISTENER.Port = port
	SOCKS_LISTENER.Cond = sync.NewCond(&SOCKS_LISTENER.Mu)
	SOCKS_LISTENER.ToSendConn = make(map[string]net.Conn)
	SOCKS_LISTENER.ConnMap = make(map[string]net.Conn)

	// Accept connections in a separate goroutine if needed
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				utils.PrintDebug("Error Accepting Connection:", err.Error())
				continue // Optionally handle error appropriately
			}
			connID := uuid.New().String()
			SOCKS_LISTENER.Mu.Lock()
			SOCKS_LISTENER.ToSendConn[connID] = conn
			SOCKS_LISTENER.Cond.Signal() // Notify anyone waiting for a connection
			SOCKS_LISTENER.Mu.Unlock()

			// Optionally log or process the connection here
			utils.PrintDebug("Accepted new connection: %s\n", connID)
		}
	}()

	return nil
}

// Wait until ToSendConn has at least one connection using sync.Cond
func waitForConnection() {
	SOCKS_LISTENER.Mu.Lock()
	// Wait until the map is not empty.
	for len(SOCKS_LISTENER.ToSendConn) == 0 {
		SOCKS_LISTENER.Cond.Wait()
	}
	SOCKS_LISTENER.Mu.Unlock()
	utils.PrintDebug("New Connection added")
}

// handleSocksEnvelope processes incoming envelopes and routes connections
func HandleSocksEnvelope(e *orstedrpc.Envelope) *orstedrpc.Envelope {
	var res orstedrpc.Envelope

	utils.PrintDebug("ConnMap", SOCKS_LISTENER.ConnMap)
	utils.PrintDebug("ToSendMap", SOCKS_LISTENER.ToSendConn)
	utils.PrintDebug("Received Envelope --> ", e.String())
	res.Data = nil
	res.Chain = e.Chain
	res.Id = e.Id
	res.Type = "socksresponse"

	if e.Type == "sockslisten" {
		utils.PrintDebug("Waiting for Proxychain connection")
		waitForConnection() // Blocks until a new connection is available

		// Retrieve the first available connection from ToSendConn:
		SOCKS_LISTENER.Mu.Lock()
		for key, conn := range SOCKS_LISTENER.ToSendConn {
			// Move connection to ConnMap for persistent tracking
			SOCKS_LISTENER.ConnMap[key] = conn
			res.Id = key
			delete(SOCKS_LISTENER.ToSendConn, key)
			utils.PrintDebug("New connection Key", res.Id)
			break // We only needed one connection, break after processing the first.
		}
		SOCKS_LISTENER.Mu.Unlock()

		return &res
	}

	// For other envelope types, retrieve the connection using the provided Id.
	SOCKS_LISTENER.Mu.Lock()
	conn, exists := SOCKS_LISTENER.ConnMap[e.Id]
	SOCKS_LISTENER.Mu.Unlock()
    if conn == nil {
        return &res
    }

	utils.PrintDebug("Received Request for ID", e.Id)
	if !exists {
		utils.PrintDebug("Connection not found")
		return &res
	}

	// Process based on envelope type.
	switch e.Type {
	case "socksread":
		utils.PrintDebug("Reading")
        size, _ := binary.Varint(e.Data)
        utils.PrintDebug("Socks Read: Reading Form buffer of size ", size)
		temp := make([]byte, size)
		n, err := conn.Read(temp)
		if err != nil {
			utils.PrintDebug("Error reading connection:", err)
            res.Type = "sockserror"
            res.Data = []byte(err.Error())
            return &res
		}
		res.Data = temp[:n]
		utils.PrintDebug("Read ", res.Data)
	case "sockswrite":
		utils.PrintDebug("Writing")
		_, err := conn.Write(e.Data)
		if err != nil {
			utils.PrintDebug("Error writing to connection:", err)
            res.Type = "sockserror"
            res.Data = []byte(err.Error())
            return &res
		}
		utils.PrintDebug("Wrote ", e.Data)
	case "socksremoteaddr":
		remote := conn.RemoteAddr()
		res.Data = []byte(remote.Network() + "//" + remote.String())
	case "sockslocaladdr":
		local := conn.LocalAddr()
		res.Data = []byte(local.Network() + "//" + local.String())
	case "socksclose":
        if conn != nil {
            conn.Close()
        }
	}

	return &res
}

