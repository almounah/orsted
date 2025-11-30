package utils

import "log"

var DEBUG bool = true

func Print(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}
