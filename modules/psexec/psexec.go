package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc/mgr"
)

func PsExec(hostname string, binPath string, fileData []byte, arguments string, serviceName string, serviceDesc string) error {
	binPath = strings.Split(binPath, ".")[0] + generateVersionID() + ".exe"
	serviceName = serviceName + generateVersionID()
	err := UploadFile(hostname, binPath, fileData)
	if err != nil {
		return err
	}

	err = StartService(hostname, binPath, arguments, serviceName, serviceDesc)
	return err
}

func generateVersionID() string {
	n := rand.Intn(90000) + 10000 
	return fmt.Sprintf("v%d", n)
}

func UploadFile(hostname string, binPath string, fileData []byte) error {
	// Combine into a UNC path: \\hostname\share\subdir\file.ext
	uncPath := fmt.Sprintf(`\\%s\%s`, hostname, binPath)
	uncPath = strings.Replace(uncPath, ":", "$", 1)

	// Create directory if it doesn't exist
	dir := filepath.Dir(uncPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Create (or overwrite) the file
	f, err := os.Create(uncPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(fileData)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func StartService(hostname string, binPath string, arguments string, serviceName string, serviceDesc string) error {
	manager, err := mgr.ConnectRemote(hostname)
	if err != nil {
		return err
	}

	service, err := manager.CreateService(serviceName, binPath, mgr.Config{
		ErrorControl:   mgr.ErrorNormal,
		BinaryPathName: binPath,
		Description:    serviceDesc,
		DisplayName:    serviceName,
		ServiceType:    windows.SERVICE_WIN32_OWN_PROCESS,
		StartType:      mgr.StartManual,
	}, arguments)

	if err != nil {
		return err
	}
	err = service.Start()
	if err != nil {
		return err
	}
	return err
}
