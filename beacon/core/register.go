package core

import (
	"encoding/json"

	"orsted/beacon/modules/gethostname"
	"orsted/beacon/modules/getos"
	"orsted/beacon/modules/netstat"
	"orsted/beacon/modules/userinfo"
	"orsted/beacon/utils"
)

// Send RegisterBeacon Message to Peer, Get Response and return Beacon Id
func RegisterBeacon(p utils.Peer) (string, error) {
	payload := map[string]interface{}{
		"os":        getos.GetOS(),
		"ip":        netstat.GetLocalIP(),
		"hostname":  gethostname.Gethostname(),
		"user":      userinfo.GetUserName(),
		"integrity": userinfo.GetUserIntegrity(),
        "transport": p.GetPeerType(),
		"chain":     nil, // or (*[]string)(nil)
	}

	jsonBytes, _ := json.Marshal(payload)

	var envelopeTosend utils.Envelope
	// Doesn't have Beacon Id so put -1
	envelopeTosend.EnvId = "-1"
	envelopeTosend.EnvData = jsonBytes
	envelopeTosend.Chain = nil
	envelopeTosend.TypeEnv = "reqregbeacon"

	jsonEnvelopeBytes, _ := json.Marshal(envelopeTosend)

	// Prepare to send to peer
	envelopePreparedBytes, err := p.PrepareRegisterBeaconData(jsonEnvelopeBytes)
    if err != nil {
        utils.Print("Error while preparing register data", err.Error())
        return "", err
    }
	utils.Print("Prepare Json Bytes for Register -> \n", string(envelopePreparedBytes))

	// Send Request
	responseBytes, err := p.SendRequest(envelopePreparedBytes)
	if err != nil {
        utils.Print("Error while sending register data network level", err.Error())
		return "", err
	}
	utils.Print("Sent Request, Received -> \n", string(responseBytes))

	// Get Response
	envelopeResponseBytes, _ := p.CleanRespFromPeerProtocol(responseBytes)
	var envelopeResponse utils.Envelope
	err = json.Unmarshal(envelopeResponseBytes, &envelopeResponse)
	if err != nil {
		utils.Print("Error while unmarsheling Register Response ", err.Error())
		utils.Print("Received Response -> \n", string(envelopeResponseBytes))
		return "", err
	}

	type temp struct {
		Id string `json:"id"`
	}
	var res temp
	err = json.Unmarshal(envelopeResponse.EnvData, &res)
	if err != nil {
		utils.Print("Error while unmarsheling Register Response EnvData ", err.Error())
		return "", err
	}
	utils.Print("Id retreived from register response successfully -> ", res.Id)

	return res.Id, nil
}
