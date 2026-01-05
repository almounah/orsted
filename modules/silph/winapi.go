package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"unicode/utf16"
	"unsafe"

	"github.com/almounah/superdeye"
	"golang.org/x/sys/windows"
)

const KeyBasicInformation = 0
const KeyFullInformation = 2
const KeyValueBasicInformation = 0
const KeyValuePartialInformation = 2
const STATUS_INVALID_HANDLE = 0xC0000008
const STATUS_BUFFER_TOO_SMALL = 0xC0000023
const STATUS_BUFFER_OVERFLOW = 0x80000005

type KEY_VALUE_BASIC_INFORMATION struct {
	TitleIndex uint32
	Type       uint32
	NameLength uint32
	Name       [1]uint16
}

type KEY_BASIC_INFORMATION struct {
	LastWriteTime int64
	TitleIndex    uint32
	NameLength    uint32
	Name          [1]uint16
}

type KEY_VALUE_PARTIAL_INFORMATION struct {
	TitleIndex uint32
	Type       uint32
	DataLength uint32
	Data       [1]byte
}

type KeyInfo struct {
	ClassName       string
	SubKeys         uint32
	MaxSubKeyLen    uint32
	MaxClassLen     uint32
	Values          uint32
	MaxValueNameLen uint32
	MaxValueLen     uint32
}

type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type OBJECT_ATTRIBUTES struct {
	Length                   uint32
	RootDirectory            uintptr
	ObjectName               *UNICODE_STRING
	Attributes               uint32
	SecurityDescriptor       uintptr
	SecurityQualityOfService uintptr
}

type KEY_FULL_INFORMATION struct {
	LastWriteTime   int64
	TitleIndex      uint32
	ClassOffset     uint32
	ClassLength     uint32
	SubKeys         uint32
	MaxNameLen      uint32
	MaxClassLen     uint32
	Values          uint32
	MaxValueNameLen uint32
	MaxValueDataLen uint32
	Class           [1]uint16
}

func NtQueryKeyInfo(hKey windows.Handle) (*KeyInfo, error) {

	var resultLength uint32
	bufferSize := uint32(512)
	buffer := make([]byte, bufferSize)

	r0, _ := superdeye.SuperdSyscall("NtQueryKey",
		uintptr(hKey),
		uintptr(KeyFullInformation),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&resultLength)),
	)

	if r0 == STATUS_BUFFER_TOO_SMALL || r0 == STATUS_BUFFER_OVERFLOW {
		bufferSize = resultLength
		buffer = make([]byte, bufferSize)

		r0, _ = superdeye.SuperdSyscall("NtQueryKey",
			uintptr(hKey),
			uintptr(KeyFullInformation),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(bufferSize),
			uintptr(unsafe.Pointer(&resultLength)),
		)
	}

	if r0 != 0 {
		return nil, fmt.Errorf("NtQueryKey failed with NTSTATUS: 0x%X", r0)
	}

	info := (*KEY_FULL_INFORMATION)(unsafe.Pointer(&buffer[0]))

	var className string
	if info.ClassLength > 0 && info.ClassOffset > 0 {
		classBytes := buffer[info.ClassOffset : info.ClassOffset+info.ClassLength]
		classU16 := make([]uint16, info.ClassLength/2)
		for i := 0; i < len(classU16); i++ {
			classU16[i] = uint16(classBytes[i*2]) | uint16(classBytes[i*2+1])<<8
		}
		className = string(utf16.Decode(classU16))
	}

	return &KeyInfo{
		ClassName:       className,
		SubKeys:         info.SubKeys,
		MaxSubKeyLen:    info.MaxNameLen,
		MaxClassLen:     info.MaxClassLen,
		Values:          info.Values,
		MaxValueNameLen: info.MaxValueNameLen,
		MaxValueLen:     info.MaxValueDataLen,
	}, nil
}

func NewUnicodeString(s string) UNICODE_STRING {
	u16 := windows.StringToUTF16(s)
	return UNICODE_STRING{
		Length:        uint16(len(s) * 2),
		MaximumLength: uint16(len(s) * 2),
		Buffer:        &u16[0],
	}
}

func FromUnicodeString(d []byte) (string, error) {
	// Credit to https://github.com/Azure/go-ntlmssp/blob/master/unicode.go for logic
	if len(d)%2 > 0 {
		return "", errors.New("Unicode (UTF 16 LE) specified, but uneven data length")
	}
	s := make([]uint16, len(d)/2)
	err := binary.Read(bytes.NewReader(d), binary.LittleEndian, &s)
	if err != nil {
		return "", err
	}
	return string(utf16.Decode(s)), nil
}

func OpenSubKeyExt(phkey windows.Handle, subkey string, opts, desiredAccess uint32) (handle windows.Handle, err error) {
	pathU16 := windows.StringToUTF16("\\Registry\\Machine\\" + subkey)
	name := UNICODE_STRING{
		Length:        uint16((len(pathU16) - 1) * 2),
		MaximumLength: uint16(len(pathU16) * 2),
		Buffer:        &pathU16[0],
	}

	attr := OBJECT_ATTRIBUTES{
		Length:     uint32(unsafe.Sizeof(OBJECT_ATTRIBUTES{})),
		ObjectName: &name,
		Attributes: windows.OBJ_CASE_INSENSITIVE,
	}

	var hKey windows.Handle
	ret, err := superdeye.SuperdSyscall("NtOpenKeyEx",
		uintptr(unsafe.Pointer(&hKey)),
		uintptr(desiredAccess),
		uintptr(unsafe.Pointer(&attr)),
		uintptr(opts),
	)

	if err != nil {
		return 0, err
	}

	runtime.KeepAlive(pathU16)
	runtime.KeepAlive(name)
	runtime.KeepAlive(attr)

	if ret != 0 {
		return 0, fmt.Errorf("NtOpenKeyEx failed: 0x%X", ret)
	}

	// For some reason sometimes hkey can be zero
	// I am doing this as a workaround. A bug maybe existant
	// in the code
	if hKey == 0 {
		return OpenSubKeyExt(phkey, subkey, opts, desiredAccess)
	}

	return hKey, nil
}

func OpenSubKey(hKey windows.Handle, subkey string) (windows.Handle, error) {
	Println("OpenSubKey called")
	return OpenSubKeyExt(hKey, subkey, 0, PermMaximumAllowed)
}

func CloseKeyHandle(hKey windows.Handle) (err error) {
	Println("Close handle is called")
	ret, err := superdeye.SuperdSyscall("NtCloseKey", uintptr(hKey))
	if ret != 0 {
		err := fmt.Errorf("NtCloseKey Failed")
		Println(err.Error())

		return err
	}
	return err
}

func QueryValue2(hKey windows.Handle, valueName string) ([]byte, uint32, error) {

	if hKey == 0 || hKey == windows.InvalidHandle {
		return nil, 0, fmt.Errorf("invalid handle: 0x%X", hKey)
	}

	valueNameU16 := windows.StringToUTF16(valueName)
	valueNameUS := UNICODE_STRING{
		Length:        uint16((len(valueNameU16) - 1) * 2),
		MaximumLength: uint16(len(valueNameU16) * 2),
		Buffer:        &valueNameU16[0],
	}

	var resultLength uint32
	bufferSize := uint32(256)

	for attempts := 0; attempts < 3; attempts++ {
		buffer := make([]byte, bufferSize)

		r0, _ := superdeye.SuperdSyscall("NtQueryValueKey",
			uintptr(hKey),
			uintptr(unsafe.Pointer(&valueNameUS)),
			uintptr(KeyValuePartialInformation),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(bufferSize),
			uintptr(unsafe.Pointer(&resultLength)),
		)

		if r0 == 0 {
			info := (*KEY_VALUE_PARTIAL_INFORMATION)(unsafe.Pointer(&buffer[0]))
			dataOffset := unsafe.Offsetof(info.Data)
			data := buffer[dataOffset : dataOffset+uintptr(info.DataLength)]
			return data, info.Type, nil
		}

		if r0 == STATUS_BUFFER_TOO_SMALL || r0 == STATUS_BUFFER_OVERFLOW {
			if resultLength > 0 && resultLength > bufferSize {
				bufferSize = resultLength
				continue
			}
			bufferSize *= 2
			continue
		}

		return nil, 0, fmt.Errorf("NtQueryValueKey failed for %s: NTSTATUS 0x%X", valueName, r0)
	}

	return nil, 0, fmt.Errorf("failed to query %s after 3 attempts", valueName)
}

func EnumValue(hKey windows.Handle, index uint32) (string, error) {

	var resultLength uint32
	bufferSize := uint32(256)
	buffer := make([]byte, bufferSize)

	r0, _ := superdeye.SuperdSyscall("NtEnumerateValueKey",
		uintptr(hKey),
		uintptr(index),
		uintptr(KeyValueBasicInformation),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&resultLength)),
	)

	if r0 == STATUS_BUFFER_TOO_SMALL || r0 == STATUS_BUFFER_OVERFLOW {
		bufferSize = resultLength
		buffer = make([]byte, bufferSize)

		r0, _ = superdeye.SuperdSyscall("NtEnumerateValueKey",
			uintptr(hKey),
			uintptr(index),
			uintptr(KeyValueBasicInformation),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(bufferSize),
			uintptr(unsafe.Pointer(&resultLength)),
		)
	}

	if r0 != 0 {
		return "", fmt.Errorf("failed to enum value %d: NTSTATUS 0x%X", index, r0)
	}

	info := (*KEY_VALUE_BASIC_INFORMATION)(unsafe.Pointer(&buffer[0]))

	nameOffset := unsafe.Offsetof(info.Name)
	nameBytes := buffer[nameOffset : nameOffset+uintptr(info.NameLength)]

	nameU16 := make([]uint16, info.NameLength/2)
	for i := 0; i < len(nameU16); i++ {
		nameU16[i] = uint16(nameBytes[i*2]) | uint16(nameBytes[i*2+1])<<8
	}

	return string(utf16.Decode(nameU16)), nil
}

func GetValueNames(hKey windows.Handle) ([]string, error) {
	Println("GetValuesName is called")
	var numValues uint32

	info, err := NtQueryKeyInfo(hKey)
	numValues = info.Values

	if err != nil {
		return nil, fmt.Errorf("failed to query key info: %w", err)
	}

	if numValues == 0 {
		return []string{}, nil
	}

	var names []string

	for i := uint32(0); i < numValues; i++ {

		valueName, err := EnumValue(
			hKey,
			i,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to enum value %d: %w", i, err)
		}

		Println("Name ", valueName)
		names = append(names, valueName)
	}

	return names, nil
}

func QueryValueString(hKey windows.Handle, valueName string) (string, error) {
	Println("QueryValue string is called")
	data, dataType, err := QueryValue2(hKey, valueName)
	if err != nil {
		return "", err
	}

	if dataType != windows.REG_SZ {
		return "", fmt.Errorf("registry value %s is not of type string (got type %d)", valueName, dataType)
	}

	if len(data) < 2 {
		return "", nil
	}

	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
	}

	return windows.UTF16ToString(u16s), nil
}

func QueryKeyInfo(hKey windows.Handle) (*KeyInfo, error) {
	Println("QueryKeyInfo is called")

	return NtQueryKeyInfo(hKey)
}

func EnumSubKey(hKey windows.Handle, index uint32) (string, error) {

	var resultLength uint32
	bufferSize := uint32(256)
	buffer := make([]byte, bufferSize)

	r0, _ := superdeye.SuperdSyscall("NtEnumerateKey",
		uintptr(hKey),
		uintptr(index),
		uintptr(KeyBasicInformation),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&resultLength)),
	)

	if r0 == STATUS_BUFFER_TOO_SMALL || r0 == STATUS_BUFFER_OVERFLOW {
		bufferSize = resultLength
		buffer = make([]byte, bufferSize)

		r0, _ = superdeye.SuperdSyscall("NtEnumerateKey",
			uintptr(hKey),
			uintptr(index),
			uintptr(KeyBasicInformation),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(bufferSize),
			uintptr(unsafe.Pointer(&resultLength)),
		)
	}

	if r0 != 0 {
		return "", fmt.Errorf("failed to enum subkey %d: NTSTATUS 0x%X", index, r0)
	}

	info := (*KEY_BASIC_INFORMATION)(unsafe.Pointer(&buffer[0]))

	nameOffset := unsafe.Offsetof(info.Name)
	nameBytes := buffer[nameOffset : nameOffset+uintptr(info.NameLength)]

	nameU16 := make([]uint16, info.NameLength/2)
	for i := 0; i < len(nameU16); i++ {
		nameU16[i] = uint16(nameBytes[i*2]) | uint16(nameBytes[i*2+1])<<8
	}

	return string(utf16.Decode(nameU16)), nil
}

func GetSubKeyNamesExt(hKey windows.Handle, subkey string, opts, desiredAccess uint32) (names []string, err error) {
	Println("GetSubKeynameExt is called")
	var hSubKey windows.Handle
	var shouldClose bool

	if subkey != "" {
		var err error
		hSubKey, err = OpenSubKeyExt(hKey, subkey, opts, desiredAccess)
		if err != nil {
			return nil, err
		}
		shouldClose = true
		defer func() {
			if shouldClose {
				CloseKeyHandle(hSubKey)
			}
		}()
	} else {
		hSubKey = hKey
	}

	info, err := QueryKeyInfo(hSubKey)
	if err != nil {
		return nil, err
	}

	names = make([]string, 0, info.SubKeys)

	for i := uint32(0); i < info.SubKeys; i++ {
		keyName, err := EnumSubKey(hSubKey, i)
		if err != nil {
			return nil, err
		}

		names = append(names, keyName)
	}

	return names, nil
}
