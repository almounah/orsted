package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"io"
	"fmt"
	"os"
)

var (
	kb       = 1024
	MAX_SIZE = 300 * kb
)

// Task of the DLL. Example Download Task
type Task struct {
	TaskId         string `json:"taskid"`
	FileToDownload string `json:"filetodownload"`

	// General Purpose
	status   string
	tasktype string

	// For tracking
	f        *os.File
	sentSize int
	fileSize int64
}

// Initialise Task with its private field
func InitialiseTask(task *Task) error {
	f, err := os.Open(task.FileToDownload)
	if err != nil {
		Println(err.Error())
		return err
	}
	task.f = f
	fs, err := task.f.Stat()
	if err != nil {
		Println(err.Error())
		return err
	}
	task.fileSize = fs.Size()

	task.tasktype = "respdownload"
	return nil
}

// Handle Task and update its status
func TaskHandler(task *Task) ([]byte, error) {
	Println("Inside the Exported Func")

	sendBuf := make([]byte, MAX_SIZE)

	n, err := task.f.Read(sendBuf)
	if err != nil && err != io.EOF {
		Println("Error reading file:", err)
		task.status = "failed"
		return nil, err
	}
	chunk := sendBuf[:n]
	task.sentSize += len(chunk) // Track how much has been sent
	if task.fileSize <= int64(task.sentSize) {
        task.status = "completed"
	} else {
        task.status = fmt.Sprintf("ongoing - %d/%d", task.sentSize, task.fileSize)
    }
	return chunk, nil
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
