package main

import "fmt"

var DEBUG = true

func Println(str... interface{}) {
    if DEBUG {
        fmt.Println("DEBUG", str)
    }
}
