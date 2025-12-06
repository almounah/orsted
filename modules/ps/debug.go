package main

import "log"

var DEBUG bool = false

func Println(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}

func Printf(format string, a ...any) {
    if DEBUG {
        log.Printf("DEBUGi "+format , a)
    }
}
