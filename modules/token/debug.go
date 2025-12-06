package main

import "log"

var DEBUG bool = false

func Println(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}
