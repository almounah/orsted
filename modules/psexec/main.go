package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"strings"
)

type Task struct {
	TaskId      string   `json:"taskid"`
	ServiceName string   `json:"servicename"`
	ServiceDesc string   `json:"servicedesc"`
	BinPath     string   `json:"binpath"`
	Hostname    string   `json:"hostname"`
	FileData    []byte   `json:"filedata"`
	Args        []string `json:"args"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "resppsexec"
	if task.Args == nil {
		task.Args = []string{""}
	}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	err = PsExec(task.Hostname, task.BinPath, task.FileData, strings.Join(task.Args, " "), task.ServiceName, task.ServiceDesc)
	if err != nil {
		Println("Error Occured --> ", err.Error())
		task.status = "failed"
		return []byte(err.Error()), err
	}
	task.status = "completed"
	return []byte("PsExec returned without errors"), err
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
