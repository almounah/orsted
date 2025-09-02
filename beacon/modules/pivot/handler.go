package pivot

import (
	"bufio"
	"encoding/json"
	"net"
	"orsted/beacon/utils"
	"strings"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	envelopeByte := scanner.Bytes()
	utils.Print("Received Envelope from child peer\n", string(envelopeByte))

	// Parse Envelope
	var receivedEnvelop utils.Envelope
	err := json.Unmarshal(envelopeByte, &receivedEnvelop)
	if err != nil {
		return
	}

	// Forward to Parent Peer
	// If it is a req from child
	//if strings.Contains(receivedEnvelop.TypeEnv, "req") {
	// TODO: Find a Way to check child peer that send and clean --> better for SMB and TCP

	// Add Current Beacon to Chain
	utils.Print("Adding Beacon to chain --> Before Adding ", receivedEnvelop.Chain)
	receivedEnvelop.Chain = append(receivedEnvelop.Chain, utils.CurrentBeaconId)
	utils.Print("Adding Beacon to chain --> After Adding ", receivedEnvelop.Chain)
	envelopeByte, _ = json.Marshal(receivedEnvelop)

	// Prepare and Forward to parent
	var tosend []byte
	switch {
	case receivedEnvelop.TypeEnv == "reqregbeacon":
		tosend, _ = utils.ParentPeer.PrepareRegisterBeaconData(envelopeByte)

	case receivedEnvelop.TypeEnv == "reqrettask":
		tosend, _ = utils.ParentPeer.PrepareRetreiveTasksData(envelopeByte)

	case receivedEnvelop.TypeEnv == "reqsenttaskres":
		tosend, _ = utils.ParentPeer.PrepareSendTaskResultsData(envelopeByte)

	case strings.HasPrefix(receivedEnvelop.TypeEnv, "sock"):
		tosend, _ = utils.ParentPeer.PrepareSocksData(envelopeByte)

	case strings.HasPrefix(receivedEnvelop.TypeEnv, "autoroute"):
		tosend, _ = utils.ParentPeer.PrepareAutorouteData(envelopeByte)

	default:
		// Optional: log or handle unknown TypeEnv
		tosend, _ = utils.ParentPeer.PrepareSendTaskResultsData(envelopeByte)
	}

	utils.Print("Will send to Peer ->\n", string(tosend))
	resp, _ := utils.ParentPeer.SendRequest(tosend)

	cleanResp, _ := utils.ParentPeer.CleanRespFromPeerProtocol(resp)
	utils.Print("Sent Request to Peer and received -> \n", string(resp))
	utils.Print("Cleaning Request from garbage -> \n", string(cleanResp))
	conn.Write(cleanResp)
	//}

	//// Forward to correct Child if it is response
	//if strings.Contains(receivedEnvelop.TypeEnv, "resp") {
	//    // TODO: Forward to next beacon
	//    childPeer := FindChild(receivedEnvelop.BeaconId)
	//
	//    childPeer.SendRequest(envelopeByte)

	//    if receivedEnvelop.TypeEnv == "respregbeacon" {
	//
	//        // check if we are the last in the chain, if it is the case, wee need to add child peer
	//        if receivedEnvelop.Chain[len(receivedEnvelop.Chain)-1] == utils.CurrentBeaconId {
	//            ip, port := conn.LocalAddr().String()
	//        }
	//
	//    }

	//}

}
