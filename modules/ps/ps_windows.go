package main

import (
	"bytes"
	"io"
	"unsafe"
	"fmt"

	"github.com/olekukonko/tablewriter"
	"golang.org/x/sys/windows"
)

func ps() ([]byte, error){
	var uReturnLen1 uint32
	var uReturnLen2 uint32
	NtQuerySystemInformation(windows.SystemProcessInformation, nil, 0, &uReturnLen1)
	var processInfo []byte
	for i := 0; i < int(uReturnLen1); i++ {
		processInfo = append(processInfo, 0)
	}

	NtQuerySystemInformation(windows.SystemProcessInformation, unsafe.Pointer(&processInfo[0]), uReturnLen1, &uReturnLen2)

	index := 0
	res := (*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&processInfo[index]))
    var outputIo bytes.Buffer
    data := [][]string{}

	for {

        pid := fmt.Sprint(res.UniqueProcessID)

		// Open Handle
		Println("Opening Handle")
		gotHandle := true
		handle, err := NtOpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(res.UniqueProcessID))
		if err != nil {
			Println("Error: ", err.Error())
			gotHandle = false
		}
		Println("Opening Handle")

		Println("Getting Process Image")
        image := fmt.Sprintf("%s", uint16PtrToUTF8Bytes(res.ImageName.Buffer, int(res.ImageName.Length)))
		if gotHandle {
			image = getProcessImage(handle)
			if image == "ERROR" || image == "" {
				image = fmt.Sprintf("%s", uint16PtrToUTF8Bytes(res.ImageName.Buffer, int(res.ImageName.Length)))
			}
		}
		Println("Got Process Image")

		// Get Owner by checking token
		Println("Getting Process owner")
		owner := ""
		if gotHandle {
			owner,_ = getProcessOwner(handle)
		}
		Println("Got Process owner")

		// Get Arch
		Println("Getting Process Arch")
		arch := ""
		if gotHandle {
			arch = getProcArch(handle)
		}
		Println("Got Process Arch")

		// Close Handle
		windows.CloseHandle(handle)

		// Append data
        data = append(data, []string{pid, image, owner, arch})

		// Update
		index = index + int(res.NextEntryOffset)
		res = (*windows.SYSTEM_PROCESS_INFORMATION)(unsafe.Pointer(&processInfo[index]))
		if res.NextEntryOffset == 0 {
			break
		}
	}
	prettyPrint(data, []string{"PID", "IMAGE", "OWNER", "ARCH"}, &outputIo)
    output := outputIo.Bytes()
	output = append([]byte("\n"), output...)
	return output, nil
}


func getProcessImage(handle windows.Handle) string {
	var uReturnLen1 uint32
	var uReturnLen2 uint32

	// Call NtQueryInformationProcess first time
	status := NtQueryInformationProcess(handle, ProcessImageFileName, nil, 0, &uReturnLen1)
	if status != nil {
		Printf("NtQueryInformationProcess failed: %s\n", status)
	}
	if uReturnLen1 == 0 {
		return "ERROR"
	}
	var processInfoUnicode []byte
	Println("Populated stuff")
	for i := 0; i < int(uReturnLen1); i++ {
		processInfoUnicode = append(processInfoUnicode, 0)
	}

	// Call NtQueryInformationProcess second time
	status = NtQueryInformationProcess(handle, ProcessImageFileName, unsafe.Pointer(&processInfoUnicode[0]), uint32(len(processInfoUnicode)), &uReturnLen2)
	if status != nil {
		Printf("NtQueryInformationProcess failed: 0x%x\n", status)
		return "ERROR"
	}

	u := (*windows.NTUnicodeString)(unsafe.Pointer(&processInfoUnicode[0]))
    image := fmt.Sprintf("%s", uint16PtrToUTF8Bytes(u.Buffer, int(u.Length)))

	Println("Native image path:", image)
	return image
}

func getProcessOwner(handle windows.Handle) (owner string, err error) {
	
	var token windows.Token
	if err = NtOpenProcessToken(handle, windows.TOKEN_QUERY, &token); err != nil {
		return
	}
	defer token.Close()

	tokenUser, err := getTokenOwner(token)
	if err != nil {
		Println("Error Getting owner", tokenUser)
		return
	}
	owner, domain, _, err := tokenUser.User.Sid.LookupAccount("")
	owner = fmt.Sprintf("%s\\%s", domain, owner)
	return
}

// getTokenOwner retrieves access token t owner account information.
func getTokenOwner(t windows.Token) (*windows.Tokenuser, error) {
	i, e := getInfo(t, windows.TokenUser, 50)
	if e != nil {
		return nil, e
	}
	return (*windows.Tokenuser)(i), nil
}

func getProcArch(pHandle windows.Handle) string{

	isWow64Process, err := IsWow64Process(pHandle)

	arch := "x86"
	if !isWow64Process {
		arch = "x64"
	}
	if err != nil {
		arch = "err"
	}
	return arch
}


func IsWow64Process(handle windows.Handle) (bool, error) {
	
	var wow64Info uintptr
	status := NtQueryInformationProcess(
		handle,
		ProcessWow64Information,
		unsafe.Pointer(&wow64Info),
		uint32(unsafe.Sizeof(wow64Info)),
		nil,
	)
	
	if status != nil {
		Printf("NtQueryInformationProcess failed: %s\n", status)
		return false, fmt.Errorf("NtQueryInformationProcess failed: %s \n", status)
	}

	return wow64Info != 0, nil
}

func getInfo(t windows.Token, class uint32, initSize int) (unsafe.Pointer, error) {
	n := uint32(initSize)
	for {
		b := make([]byte, n)
		e := NtQueryInformationToken(t, class, &b[0], uint32(len(b)), &n)
		if e == nil {
			return unsafe.Pointer(&b[0]), nil
		}
		if e != windows.ERROR_INSUFFICIENT_BUFFER {
			return nil, e
		}
		if n <= uint32(len(b)) {
			return nil, e
		}
	}
}

func prettyPrint(data [][]string, headers []string, out io.Writer) {
	table := tablewriter.NewWriter(out)
	table.Header(headers)
	for _, v := range data {
		table.Append(v)
	}
	table.Render()
}

func uint16PtrToUTF8Bytes(ptr *uint16, length int) []byte {
	byteSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptr)), length)
	var res []byte
	for i := 0; i < len(byteSlice); i++ {
		if i%2 == 0 {
			res = append(res, byteSlice[i])
		}
	}

	return res
}
