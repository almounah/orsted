package getos

import "runtime"

func GetOS() (string) {
    return runtime.GOOS
}
