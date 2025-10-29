package autoroute

import (
	"encoding/binary"
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

	utils.PrintDebug("Received envelope", e)
	utils.PrintDebug("Received envelope Id", e.Id)
	var r *Route
	for _, c := range ROUTE_LIST {
		utils.PrintDebug("BeaconId from Route List", c.BeaconId)
		if c.BeaconId == e.Id {
			r = c
		}
	}

	if r == nil {
		res.Data = []byte("Cannot find ID")
		res.Type = "socskerror"
		return &res
	}

	utils.PrintDebug("Locking Proxy")
	r.ProxyMu.Lock()
	for r.ProxyMu == nil {
		r.ProxyCond.Wait()
	}
	r.ProxyMu.Unlock()
	utils.PrintDebug("UnLocking Proxy")

	utils.PrintDebug("Locking Agent")
	r.AgentMu.Lock()
	for r.Agent == nil {
		r.AgentCond.Wait()
	}
	r.AgentMu.Unlock()
	utils.PrintDebug("UnLocking Agent")

	SERVER_CONN := r.Agent
	// Process based on envelope type.
	switch e.Type {
	case "autorouteread":
		utils.PrintDebug("Reading")
		size, _ := binary.Varint(e.Data)
		utils.PrintDebug("Autoroute Read: Reading Form buffer of size ", size)
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

	utils.PrintDebug("------------------")
	utils.PrintDebug("Received Envelope ", e.Type, e.Data)
	utils.PrintDebug("Response Envelope ", res.Type, res.Data)
	utils.PrintDebug("------------------")

	return &res
}
