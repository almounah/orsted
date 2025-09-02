package memorymodule

import (
	"errors"
	"fmt"
	"runtime"
	"unsafe"

	"orsted/beacon/modules/memorymodule/utilslinux"
	"orsted/beacon/modules/memorymodule/utilslinux/dl"
)

type HMEMORYMODULE struct {
	Name   string
	Handle *dl.DL
}

var LoadedMemoryModule []HMEMORYMODULE

func LoadMemoryModule(moduleName string, data []byte) (stdoutbyte []byte, stderrbyte []byte, err error) {
	fd, err := utils.Open_ramfs(moduleName)
	if err != nil {
		return nil, nil, err
	}

	err = utils.Write(fd, data)
	if err != nil {
		return nil, nil, err
	}

	var filePath string
	kernelVersion, err := utils.GetKernelVersion()

	if utils.IsKernelGreaterThan(kernelVersion, 3, 17) {
		filePath = fmt.Sprintf("/dev/shm/%s", moduleName)
	} else {
		filePath = fmt.Sprintf("/proc/self/fd/%d", fd)
	}

	handle, err := dl.Open(filePath, dl.RTLD_LAZY)
	LoadedMemoryModule = append(LoadedMemoryModule, HMEMORYMODULE{moduleName, handle})
	if err != nil {
		return nil, nil, err
	}
	return []byte("Loaded Linux Module Successfully"), nil, err
}

func ExecuteFunctionInModule(moduleName string, functionName string, argument ...uintptr) (stdout []byte, stderr []byte, err error) {
	for _, module := range LoadedMemoryModule {
		if module.Name == moduleName {
			handle := module.Handle

			// Declaring common variables
			var stdoutp, stderrp byte
			var stdoutPtr, stderrPtr *byte
			var stdoutSize, stderrSize int
			stdoutPtr = &stdoutp
			stderrPtr = &stderrp

			pinner := runtime.Pinner{}
			defer pinner.Unpin()
			pinner.Pin(&stdoutPtr)  // Pinning **byte (address of stdoutPtr)
			pinner.Pin(&stdoutSize) // Pinning *int
			pinner.Pin(&stderrPtr)  // Pinning **byte (address of stderrPtr)
			pinner.Pin(&stderrSize) // Pinning *int

			switch functionName {
			case "RegisterTask":
				var RegisterTask func(taskJson *byte, taskJsonSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &RegisterTask)
				arg0 := (*byte)(unsafe.Pointer(argument[0]))
				arg1 := (*int)(unsafe.Pointer(argument[1]))
				pinner.Pin(arg0) // Pinning *int
				pinner.Pin(arg1) // Pinning *int
				RegisterTask(arg0, arg1, &stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
			case "GetNextByteChunck":
				var GetNextByteChunck func(taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &GetNextByteChunck)
				arg0 := (*byte)(unsafe.Pointer(argument[0]))
				arg1 := (*int)(unsafe.Pointer(argument[1]))
				pinner.Pin(arg0) // Pinning *int
				pinner.Pin(arg1) // Pinning *int
				GetNextByteChunck(arg0, arg1, &stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
			case "GetTaskIdList":
				var GetTaskIdList func(stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &GetTaskIdList)
				GetTaskIdList(&stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
			case "GetTaskStatus":
				var GetTaskStatus func(taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &GetTaskStatus)
				arg0 := (*byte)(unsafe.Pointer(argument[0]))
				arg1 := (*int)(unsafe.Pointer(argument[1]))
				pinner.Pin(arg0) // Pinning *int
				pinner.Pin(arg1) // Pinning *int
				GetTaskStatus(arg0, arg1, &stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
			case "GetTaskType":
                var GetTaskType func (taskIdByte *byte, taskIdSize *int, stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &GetTaskType)
				arg0 := (*byte)(unsafe.Pointer(argument[0]))
				arg1 := (*int)(unsafe.Pointer(argument[1]))
				pinner.Pin(arg0) // Pinning *int
				pinner.Pin(arg1) // Pinning *int
				GetTaskType(arg0, arg1, &stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
			case "CleanDoneTask":
                var CleanDoneTask func (stdout **byte, stdoutSize *int, stderr **byte, stderrSize *int)
				handle.Sym(functionName, &CleanDoneTask)
				CleanDoneTask(&stdoutPtr, &stdoutSize, &stderrPtr, &stderrSize)
            default:
                return nil, nil, errors.New("Function Name not found")


			}

			stdoutC := unsafe.Slice(stdoutPtr, stdoutSize)
			stderrC := unsafe.Slice(stderrPtr, stderrSize)

			stdout = make([]byte, stdoutSize)
			copy(stdout, stdoutC)
			stderr = make([]byte, stderrSize)
			copy(stderr, stderrC)

			return []byte(stdout), []byte(stderr), nil
		}
	}

	return nil, nil, errors.New("Module Not Found be sure to load it")
}
