package main

import "fmt"

var DEBUG = false

func Println(str... interface{}) {
    if DEBUG {
        fmt.Println("DEBUG", str)
    }
}
