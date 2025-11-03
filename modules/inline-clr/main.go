package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"sync"
)

type Task struct {
	TaskId     string   `json:"taskid"`
	TaskAction string   `json:"taskaction"`
	Assembly   []byte   `json:"assembly"`
	AssemblyName   string   `json:"assemblyname"`
	Args       []string `json:"args"`

	// General Purpose
	status   string
	tasktype string

	mu *sync.Mutex
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respinlineexecute"
	task.mu = &sync.Mutex{}
    if task.Args == nil {
        task.Args = []string{" "}
    }
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
    Println("Task Action is ---->", task.TaskAction)
	switch task.TaskAction {
	case "start-clr":
        stdout, err := startCLR()
        task.status = "completed"
        return stdout, err
    case "load-assembly":
        stdout, err := loadAssembly(task.AssemblyName, task.Assembly)
        task.status = "completed"
        return stdout, err
    case "invoke-assembly":
        stdout, err := invokeAssembly(task.AssemblyName, task.Args)
        task.status = "completed"
        return stdout, err
    case "list-assemblies":
        stdout, err := listAssemblies()
        task.status = "completed"
        return stdout, err
	}
	if err != nil {
		Println("Fail")
	}
	Println("Done Executing")
	Println("Result From DLL --->")
	Println(string(stdout))
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
