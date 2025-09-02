package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
)

type Task struct {
	TaskId     string `json:"taskid"`

	// General Purpose
	status   string
	tasktype string

}

func InitialiseTask(task *Task) error {
	task.tasktype = "respps"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	stdout, err = ps()
    if err != nil {
        fmt.Println("Error Occured --> ", err.Error())
		task.status = "failed"
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
