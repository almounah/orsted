package autoroute

import (
	"encoding/binary"
	"fmt"
	"orsted/protobuf/orstedrpc"
	"orsted/server/utils"

)

func HandleAutorouteMessage(e *orstedrpc.Envelope) *orstedrpc.Envelope {
	var res orstedrpc.Envelope

	res.Data = nil
	res.Chain = e.Chain
	res.Id = e.Id
	res.Type = "autorouteresponse"

	if e.Type == "autoroutelisten" {
		res.Data = []byte("Not Supported")

	}

	fmt.Println("Received envelope", e)
	fmt.Println("Received envelope Id", e.Id)
	var r *Route
	for _, c := range ROUTE_LIST {
		fmt.Println("BeaconId from Route List", c.BeaconId)
		if c.BeaconId == e.Id {
			r = c
		}
	}

	if r == nil {
		res.Data = []byte("Cannot find ID")
		res.Type = "socskerror"
		return &res
	}

	fmt.Println("Locking Proxy")
	r.ProxyMu.Lock()
	for r.ProxyMu == nil {
		r.ProxyCond.Wait()
	}
	r.ProxyMu.Unlock()
	fmt.Println("UnLocking Proxy")

	fmt.Println("Locking Agent")
	r.AgentMu.Lock()
	for r.Agent == nil {
		r.AgentCond.Wait()
	}
	r.AgentMu.Unlock()
	fmt.Println("UnLocking Agent")

	SERVER_CONN := r.Agent
	// Process based on envelope type.
	switch e.Type {
	case "autorouteread":
		utils.PrintDebug("Reading")
		size, _ := binary.Varint(e.Data)
		utils.PrintDebug("Socks Read: Reading Form buffer of size ", size)
		temp := make([]byte, size)
		n, err := SERVER_CONN.Read(temp)
		if err != nil {
			utils.PrintDebug("Error reading connection:", err)
			res.Type = "sockserror"
			res.Data = []byte(err.Error())
		}
		res.Data = temp[:n]
		utils.PrintDebug("Read ", res.Data)
	case "autoroutewrite":
		utils.PrintDebug("Writing")
		n, err := SERVER_CONN.Write(e.Data)
		if err != nil {
			utils.PrintDebug("Error writing to connection:", err)
			res.Type = "sockserror"
			res.Data = []byte(err.Error())
		}
		res.Data = []byte{byte(n)}
		utils.PrintDebug("Wrote ", e.Data)
	case "autorouteremoteaddr":
		remote := SERVER_CONN.RemoteAddr()
		res.Data = []byte(remote.Network() + "//" + remote.String())
	case "autoroutelocaladdr":
		local := SERVER_CONN.LocalAddr()
		res.Data = []byte(local.Network() + "//" + local.String())
	case "autorouteclose":
		if SERVER_CONN != nil {
			SERVER_CONN.Close()
		}
	}

	fmt.Println("------------------")
	fmt.Println("Received Envelope ", e.Type, e.Data)
	fmt.Println("Response Envelope ", res.Type, res.Data)
	fmt.Println("------------------")

	return &res
}
