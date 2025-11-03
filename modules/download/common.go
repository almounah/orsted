package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

// List of the Task Registered in Beacon
var TASK_LIST []*Task

// Add Task to DLL/.so Memory
// taskJson is the json of the task
//
//export RegisterTask
func RegisterTask(taskJson *byte, taskJsonSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	var task Task
	var stdoutBytes []byte
	var stderrBytes []byte

	Println("Inside the Exported RegosterTask Func")

	// Unmarshall Task Json
	taskJsonByteArray := unsafe.Slice(taskJson, *taskJsonSize)
	err := json.Unmarshal(taskJsonByteArray, &task)
	if err != nil {
		stdoutBytes = []byte("")
		stderrBytes = []byte(err.Error())
		*stdoutSize = len(stdoutBytes)
		cStdout := C.malloc(C.size_t(len(stdoutBytes)))
		copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
		*stdout = (*byte)(cStdout)

		*stderrSize = len(stderrBytes)
		cStderr := C.malloc(C.size_t(len(stderrBytes)))
		copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
		*stderr = (*byte)(cStderr)
        return
	}

	Println(task)

    // Initialise Task
    err = InitialiseTask(&task)

	if err != nil {
		stdoutBytes = []byte("")
		stderrBytes = []byte(err.Error())
	} else {
		stdoutBytes = []byte("Task Register Successufully")
		stderrBytes = []byte("")
        TASK_LIST = append(TASK_LIST, &task)
	}

	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

// Do what the DLL Needs and send the stdout and stderr
//
//export GetNextByteChunck
func GetNextByteChunck(taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	taskId := string(unsafe.Slice(taskIdByte, *taskIdSize))
	Println("Executin task --->", taskId)
	var task *Task = &Task{}
	for i := 0; i < len(TASK_LIST); i++ {
		if TASK_LIST[i].TaskId == taskId {
			task = TASK_LIST[i]
			break
		}
	}
	stdoutBytes, err := TaskHandler(task)
	if err != nil {
		Println(err.Error())
	}
	Println("Task HAndler Called with --->", stdoutBytes)

	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	stderrBytes := []byte("")
	if err != nil {
		stderrBytes = []byte(err.Error())
	}
	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

//export GetTaskIdList
func GetTaskIdList(stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	var result []string

	for i := 0; i < len(TASK_LIST); i++ {
		result = append(result, TASK_LIST[i].TaskId)
	}
	Println("Will Send Result Back to Hist ---->")
	Println(result)

	stdoutBytes, _ := json.Marshal(result)
	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	stderrBytes := []byte("")
	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

//export GetTaskStatus
func GetTaskStatus(taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	taskId := string(unsafe.Slice(taskIdByte, *taskIdSize))
	var task *Task = &Task{}
	for i := 0; i < len(TASK_LIST); i++ {
		if TASK_LIST[i].TaskId == taskId {
			task = TASK_LIST[i]
			break
		}
	}
	stdoutBytes := ComputeTaskStatus(*task)

	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	stderrBytes := []byte("")
	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

//export GetTaskType
func GetTaskType(taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	taskId := string(unsafe.Slice(taskIdByte, *taskIdSize))
	var task *Task = &Task{}
	for i := 0; i < len(TASK_LIST); i++ {
		if TASK_LIST[i].TaskId == taskId {
			task = TASK_LIST[i]
			break
		}
	}
	stdoutBytes := ComputeTaskType(*task)

	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	stderrBytes := []byte("")
	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

//export CleanDoneTask
func CleanDoneTask(stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int) {
	newTaskList := []*Task{}
	for i := 0; i < len(TASK_LIST); i++ {
		if !IsTaskDone(*TASK_LIST[i]) {
			newTaskList = append(newTaskList, TASK_LIST[i])
		}
	}
	TASK_LIST = newTaskList

	stdoutBytes := []byte("")
	*stdoutSize = len(stdoutBytes)
	cStdout := C.malloc(C.size_t(len(stdoutBytes)))
	copy((*[1 << 30]byte)(cStdout)[:], stdoutBytes)
	*stdout = (*byte)(cStdout)

	stderrBytes := []byte("")
	*stderrSize = len(stderrBytes)
	cStderr := C.malloc(C.size_t(len(stderrBytes)))
	copy((*[1 << 30]byte)(cStderr)[:], stderrBytes)
	*stderr = (*byte)(cStderr)
}

//export FreeMem
func FreeMem(ptr *byte) {
	C.free(unsafe.Pointer(ptr)) // Free using the SAME C runtime as the DLL
}

func init() {
}

func main() {
	// go build -o dllmain.dll -buildmode=c-shared cmd/dllFortest/main.go
}
