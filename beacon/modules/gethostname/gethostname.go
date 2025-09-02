package gethostname

import "os"

func Gethostname() (string) {

    h, _ := os.Hostname()
    return h

}
