package core

import (
	"encoding/base64"
	"encoding/json"
	//"fmt"
	"orsted/beacon/utils"
)

// Send Result of a Task by contacting peer
// restype can be respdownload to instruct server to download file 
// instead of just savin the repsonse
func SendTaskResult(p utils.Peer, taskId string, state string, restype string, response []byte) error {
	payload := map[string]string{
		"taskId":   taskId,
		"state":    state,
		"response": base64.StdEncoding.EncodeToString(response),
	}

	jsonBytes, _ := json.Marshal(payload)

	// Creating Envelop
	var envelopeTosend utils.Envelope
	envelopeTosend.EnvId = taskId
	envelopeTosend.EnvData = jsonBytes
	envelopeTosend.Chain = nil
	envelopeTosend.TypeEnv = restype

	jsonEnvelopeBytes, _ := json.Marshal(envelopeTosend)

	// Prepare to send to peer
	envelopePreparedBytes, _ := p.PrepareSendTaskResultsData(jsonEnvelopeBytes)

	// Send Request
	responseBytes, _ := p.SendRequest(envelopePreparedBytes)

	// Get Response
	envelopeResponseBytes, _ := p.CleanRespFromPeerProtocol(responseBytes)
	var envelopeResponse utils.Envelope
	err := json.Unmarshal(envelopeResponseBytes, &envelopeResponse)
	if err != nil {
		//utils.Print("Error Sending task result", err.Error())
		return err
	}
	utils.Print("Sending Task Result")
	utils.Print(string(response))

	// TODO: Check if response is correct

	return nil
}


// Retreive Beacon Tasks by contacting peer
func RetreiveTask(p utils.Peer, id string) (*utils.Tasks, error) {
	payload := map[string]string{
		"id": id,
	}
	jsonBytes, _ := json.Marshal(payload)

	// Creating Envelop
	var envelopeTosend utils.Envelope
	envelopeTosend.EnvId = utils.CurrentBeaconId
	envelopeTosend.EnvData = jsonBytes
	envelopeTosend.Chain = nil
	envelopeTosend.TypeEnv = "reqrettask"

	jsonEnvelopeBytes, _ := json.Marshal(envelopeTosend)

	// Prepare to send to peer
	envelopePreparedBytes, _ := p.PrepareRetreiveTasksData(jsonEnvelopeBytes)

	// Send Request
	responseBytes, err := p.SendRequest(envelopePreparedBytes)
	if err != nil {
		utils.Print("Error while getting task list at Nework level -> ", err.Error())
		utils.Print("Sent following request to get task ->", string(envelopePreparedBytes))
		return nil, err
	}

	// Get Response
	envelopeResponseBytes, _ := p.CleanRespFromPeerProtocol(responseBytes)
	var envelopeResponse utils.Envelope
	err = json.Unmarshal(envelopeResponseBytes, &envelopeResponse)
	if err != nil {
		utils.Print("Error while getting task list -> ", err.Error())
		utils.Print("Sent following request to get task ->", string(envelopePreparedBytes))
		return nil, err
	}

	var taskList utils.Tasks
	// fmt.Println("You fucked here")
	err = json.Unmarshal(envelopeResponse.EnvData, &taskList)
	if err != nil {
		return nil, err
	}
	// fmt.Println("You fucked here no")
	return &taskList, nil
}
