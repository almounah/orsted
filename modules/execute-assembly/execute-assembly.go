package main

import (
	"errors"
	"execute-assembly/utils"
	"fmt"
)

func executeAssembly(shellcode []byte, method string, processImage string) (stdout []byte, stderr []byte, err error) {
    switch method {
    case "1":
        fmt.Println("Will Run early bird")
        return utils.RunEarlyBird(shellcode, processImage)
    }
    return nil, nil, errors.New("Method number not found")
}

func executeAssembly1(shellcode []byte, processImage string) (stdout []byte, stderr []byte, err error) {
    return utils.RunEarlyBird(shellcode, processImage)
}
