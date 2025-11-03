package main

import (
	"errors"
	"execute-assembly/utils"
)

func executeAssembly(shellcode []byte, method string, processImage string) (stdout []byte, stderr []byte, err error) {
    switch method {
    case "1":
        Println("Will Run early bird")
        return utils.RunEarlyBird(shellcode, processImage)
    }
    return nil, nil, errors.New("Method number not found")
}

func executeAssembly1(shellcode []byte, processImage string) (stdout []byte, stderr []byte, err error) {
    return utils.RunEarlyBird(shellcode, processImage)
}
