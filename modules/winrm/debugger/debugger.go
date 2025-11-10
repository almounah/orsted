package debugger

import "log"

var DEBUG bool = true

func Println(str... interface{}) {
    if DEBUG {
        log.Println("DEBUG", str)
    }
}
