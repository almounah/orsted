package utils

import (
	"log"

)

var DEBUG bool = false
var INFO bool = true

func PrintDebug(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}

func PrintInfo(str... interface{}) {
    if INFO {
        log.Println("INFO", str)
    }
}
