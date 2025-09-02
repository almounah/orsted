package main

import (
	"errors"
	"strings"
	"sync"

	clr "inline-clr/go-buena-clr"
)

var runtimeHost *clr.ICORRuntimeHost

var assemblies = make(map[string]assembly)

var redirected bool

var Mutex sync.Mutex

type assembly struct {
	name       string
	methodInfo *clr.MethodInfo
}

func startCLR() ([]byte, error) {
	runtimeHosttemp, err := clr.LoadCLR("v4")
	runtimeHost = runtimeHosttemp
	err = clr.RedirectStdoutStderr()

	var res []byte
	if err != nil {
		res = []byte("Failed to Start CLR")
	} else {
		res = []byte("Started CLR Successfully")
	}
	appDomain, err := clr.GetAppDomain(runtimeHost)
	clr.PatchSysExit(appDomain)
	return res, err
}

func loadAssembly(assemblyName string, assemblyByte []byte) ([]byte, error) {
    if runtimeHost == nil {
		res := []byte("Failed to Load Assembly")
		return res, errors.New("CLR Not Started")
    }
	var a assembly
	a.name = strings.ToLower(assemblyName)
	methodInfo, err := clr.LoadAssembly(runtimeHost, assemblyByte)
	a.methodInfo = methodInfo
	if err != nil {
		res := []byte("Failed to LoadAssembly")
		return res, err
	}
	assemblies[a.name] = a
	return []byte("Loaded Assembly Successfully"), err
}

func invokeAssembly(assemblyName string, args []string) ([]byte, error) {
    if runtimeHost == nil {
		res := []byte("Failed to InvokeAssembly")
		return res, errors.New("CLR Not Started")
    }
	var isLoaded bool
	var a assembly
    if args == nil {
        args = []string{" "}
    }
	for _, v := range assemblies {
		if v.name == strings.ToLower(assemblyName) {
			isLoaded = true
			a = v
		}
	}
	if !isLoaded {
		res := []byte("Assembly Not Loaded")
		return res, errors.New("Assembly Not Loaded")
	}
	Mutex.Lock()
	Stdout, Stderr := clr.InvokeAssembly(a.methodInfo, args)
	Mutex.Unlock()

	return []byte(Stdout + "\n" + Stderr), nil
}

func listAssemblies() ([]byte, error) {
    result := "Loaded Assemblies are:\n"
    result += "----------------------\n"
	for _, v := range assemblies {
		result += strings.ToLower(v.name)
        result += "\n"
	}
    return []byte(result), nil
}
