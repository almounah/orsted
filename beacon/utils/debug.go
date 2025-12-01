package utils

import "log"

var DEBUG bool = false

func Print(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}
