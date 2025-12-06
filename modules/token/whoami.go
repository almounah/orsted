package main

import (
	"bytes"
	"encoding/binary"
	"unsafe"
	"fmt"

	"golang.org/x/sys/windows"
)

func whoami() ([]byte, error) {
	// Process
	tProc := windows.GetCurrentProcessToken()
	result, err := GetTokenDescriptionString(tProc)
	if err != nil {
		return result, err
	}

	tthread := windows.GetCurrentThreadToken()
	res2, err := GetTokenDescriptionString(tthread)

	returnRes := fmt.Sprintf("%s \n\nThread Token for thread ID %d:\n %s", string(result), windows.GetCurrentThreadId(), string(res2))
	return []byte(returnRes), nil

}

func GetTokenDescriptionString(t windows.Token) ([]byte, error) {
	username, err := GetTokenUsername(t)

	statProc, err := GetTokenStats(t)
	if err != nil {
		return []byte("Error Getting Token Stats " + err.Error()), err
	}

	privs, err := GetTokenPrivileges(t)
	if err != nil {
		return []byte("Error Getting Token Privileges " + err.Error()), err
	}

	groups, err := GetTokenGroups(t)
	if err != nil {
		return []byte("Error Getting Token Groups " + err.Error()), err
	}

	results := ""
	results += fmt.Sprintf("\nToken Type: %s\n", TokenTypeToString(statProc.TokenType))
	results += fmt.Sprintf("----------------\n")
	results += fmt.Sprintf("----------------\n")
	results += fmt.Sprintf("\tUser: %s\n", username)
	results += fmt.Sprintf("\tToken ID: 0x%X\n", statProc.TokenId.LowPart)
	results += fmt.Sprintf("\tLogon ID: 0x%X\n\n", statProc.AuthenticationId.LowPart)
	results += fmt.Sprintf("\tPrivilege Count: %d\n", statProc.PrivilegeCount)
	results += fmt.Sprintf("\t===================\n")
	for _, priv := range privs {
		results += fmt.Sprintf("\tPrivilege: %s, Attribute: %s\n", PrivilegeToString(priv.Luid), PrivilegeAttributeToString(priv.Attributes))
	}
	results += fmt.Sprintf("\n")
	results += fmt.Sprintf("\tGroup Count: %d\n", statProc.GroupCount)
	results += fmt.Sprintf("\t===============\n")
	for _, group := range groups {
		results += fmt.Sprintf("\tGroup: %s\n", GetGroupNameFromSid(group.Sid))
	}
	results += fmt.Sprintf("\n")
	results += fmt.Sprintf("\tType: %s\n", TokenTypeToString(statProc.TokenType))
	results += fmt.Sprintf("\tImpersonation Level: %s\n", ImpersonationToString(statProc.ImpersonationLevel))
	pLevel, err := GetTokenIntegrityLevel(t)
	if err != nil {
		return []byte("Error Getting Token Description: " + err.Error()), err
	}
	results += fmt.Sprintf("\tIntegrity Level: %s\n", pLevel)
	return []byte(results), nil

}

func GetTokenUsername(token windows.Token) (username string, err error) {
	user, err := token.GetTokenUser()
	if err != nil {
		return "", fmt.Errorf("there was an error calling GetTokenUser(): %s", err)
	}

	account, domain, _, err := user.User.Sid.LookupAccount("")
	if err != nil {
		return "", fmt.Errorf("there was an error calling SID.LookupAccount(): %s", err)
	}

	username = fmt.Sprintf("%s\\%s", domain, account)
	return
}

func GetTokenStats(token windows.Token) (tokenStats TOKEN_STATISTICS, err error) {
	// Determine the size needed for the structure
	// BOOL GetTokenInformation(
	//  [in]            HANDLE                  TokenHandle,
	//  [in]            TOKEN_INFORMATION_CLASS TokenInformationClass,
	//  [out, optional] LPVOID                  TokenInformation,
	//  [in]            DWORD                   TokenInformationLength,
	//  [out]           PDWORD                  ReturnLength
	//);
	var returnLength uint32
	err = NtQueryInformationToken(token, windows.TokenStatistics, nil, 0, &returnLength)
	if returnLength == 0 {
		err = fmt.Errorf("there was an error calling first windows.GetTokenInformation in  GetTokenStat: %s", err)
		return
	}

	// Make the call with the known size of the object
	info := bytes.NewBuffer(make([]byte, returnLength))
	var returnLength2 uint32
	err = NtQueryInformationToken(token, windows.TokenStatistics, &info.Bytes()[0], returnLength, &returnLength2)
	if err != nil {
		err = fmt.Errorf("there was an error calling second windows.GetTokenInformation in GetTokenStat: %s", err)
		return
	}

	err = binary.Read(info, binary.LittleEndian, &tokenStats)
	if err != nil {
		err = fmt.Errorf("there was an error reading binary into the TOKEN_STATISTICS structure: %s", err)
		return
	}
	return
}

func GetTokenIntegrityLevel(token windows.Token) (string, error) {
	var info byte
	var returnedLen uint32
	// Call the first time to get the output structure size
	err := NtQueryInformationToken(token, windows.TokenIntegrityLevel, &info, 0, &returnedLen)
	if returnedLen == 0 {
		return "", fmt.Errorf("there was an error calling first windows.GetTokenInformation in GetTokenIntegrityLevel: %s", err)
	}

	// Knowing the structure size, call again
	TokenIntegrityInformation := bytes.NewBuffer(make([]byte, returnedLen))
	err = NtQueryInformationToken(token, windows.TokenIntegrityLevel, &TokenIntegrityInformation.Bytes()[0], returnedLen, &returnedLen)
	if err != nil {
		return "", fmt.Errorf("there was an error calling second windows.GetTokenInformationi in GetTokenINtegrityLevel: %s", err)
	}

	// Read the buffer into a byte slice
	bLabel := make([]byte, returnedLen)
	err = binary.Read(TokenIntegrityInformation, binary.LittleEndian, &bLabel)
	if err != nil {
		return "", fmt.Errorf("there was an error reading bytes for the token integrity level: %s", err)
	}

	// Integrity level is in the Attributes portion of the structure, a DWORD, the last four bytes
	// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-token_mandatory_label
	// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-sid_and_attributes
	integrityLevel := binary.LittleEndian.Uint32(bLabel[returnedLen-4:])
	return integrityLevelToString(integrityLevel), nil
}

func integrityLevelToString(level uint32) string {
	switch level {
	case 0x00000000: // SECURITY_MANDATORY_UNTRUSTED_RID
		return "Untrusted"
	case 0x00001000: // SECURITY_MANDATORY_LOW_RID
		return "Low"
	case 0x00002000: // SECURITY_MANDATORY_MEDIUM_RID
		return "Medium"
	case 0x00002100: // SECURITY_MANDATORY_MEDIUM_PLUS_RID
		return "Medium High"
	case 0x00003000: // SECURITY_MANDATORY_HIGH_RID
		return "High"
	case 0x00004000: // SECURITY_MANDATORY_SYSTEM_RID
		return "System"
	case 0x00005000: // SECURITY_MANDATORY_PROTECTED_PROCESS_RID
		return "Protected Process"
	default:
		return fmt.Sprintf("Uknown integrity level: %d", level)
	}
}

func TokenTypeToString(tokenType uint32) string {
	switch tokenType {
	case windows.TokenPrimary:
		return "Primary"
	case windows.TokenImpersonation:
		return "Impersonation"
	default:
		return fmt.Sprintf("unknown TOKEN_TYPE: %d", tokenType)
	}
}

// ImpersonationToString converts a SECURITY_IMPERSONATION_LEVEL uint32 value to it's associated string
func ImpersonationToString(level uint32) string {
	switch level {
	case windows.SecurityAnonymous:
		return "Anonymous"
	case windows.SecurityIdentification:
		return "Identification"
	case windows.SecurityImpersonation:
		return "Impersonation"
	case windows.SecurityDelegation:
		return "Delegation"
	default:
		return fmt.Sprintf("unknown SECURITY_IMPERSONATION_LEVEL: %d", level)
	}
}

func GetTokenPrivileges(token windows.Token) (privs []windows.LUIDAndAttributes, err error) {
	// Get the privileges and attributes
	// Call to get structure size
	var returnedLen uint32
	err = NtQueryInformationToken(token, windows.TokenPrivileges, nil, 0, &returnedLen)
	if returnedLen == 0 {
		err = fmt.Errorf("there was an error calling first windows.GetTokenInformation in GetTokenPrivilege: %s", err)
		return
	}

	// Call again to get the actual structure
	info := bytes.NewBuffer(make([]byte, returnedLen))
	err = NtQueryInformationToken(token, windows.TokenPrivileges, &info.Bytes()[0], returnedLen, &returnedLen)
	if err != nil {
		err = fmt.Errorf("there was an error calling second windows.GetTokenInformation in GetTokenPrivilege: %s", err)
		return
	}

	var privilegeCount uint32
	err = binary.Read(info, binary.LittleEndian, &privilegeCount)
	if err != nil {
		err = fmt.Errorf("there was an error reading TokenPrivileges bytes to privilegeCount: %s", err)
		return
	}

	// Read in the LUID and Attributes
	for i := 1; i <= int(privilegeCount); i++ {
		var priv windows.LUIDAndAttributes
		err = binary.Read(info, binary.LittleEndian, &priv)
		if err != nil {
			err = fmt.Errorf("there was an error reading LUIDAttributes to bytes: %s", err)
			return
		}
		privs = append(privs, priv)
	}
	return
}

// PrivilegeAttributeToString converts a privilege attribute integer to a string
func PrivilegeAttributeToString(attribute uint32) string {
	// https://docs.microsoft.com/en-us/windows/win32/api/winnt/ns-winnt-token_privileges
	switch attribute {
	case 0x00000000:
		return ""
	case 0x00000001:
		return "SE_PRIVILEGE_ENABLED_BY_DEFAULT"
	case 0x00000002:
		return "SE_PRIVILEGE_ENABLED"
	case 0x00000001 | 0x00000002:
		return "SE_PRIVILEGE_ENABLED_BY_DEFAULT,SE_PRIVILEGE_ENABLED"
	case 0x00000004:
		return "SE_PRIVILEGE_REMOVED"
	case 0x80000000:
		return "SE_PRIVILEGE_USED_FOR_ACCESS"
	case 0x00000001 | 0x00000002 | 0x00000004 | 0x80000000:
		return "SE_PRIVILEGE_VALID_ATTRIBUTES"
	default:
		return fmt.Sprintf("Unknown SE_PRIVILEGE_ value: 0x%X", attribute)
	}
}

// PrivilegeToString converts a LUID to it's string representation
func PrivilegeToString(priv windows.LUID) string {
	p, err := LookupPrivilegeName(priv)
	if err != nil {
		return err.Error()
	}
	return p
}

var Advapi32 = windows.NewLazySystemDLL("Advapi32.dll")

func LookupPrivilegeName(luid windows.LUID) (privilege string, err error) {
	lookupPrivilegeNameW := Advapi32.NewProc("LookupPrivilegeNameA")

	// BOOL LookupPrivilegeNameW(
	//  [in, optional]  LPCWSTR lpSystemName,
	//  [in]            PLUID   lpLuid,
	//  [out, optional] LPWSTR  lpName,
	//  [in, out]       LPDWORD cchName
	//);

	// Call to determine the size
	var cchName uint32
	ret, _, err := lookupPrivilegeNameW.Call(0, uintptr(unsafe.Pointer(&luid)), 0, uintptr(unsafe.Pointer(&cchName)))
	if err != windows.ERROR_INSUFFICIENT_BUFFER {
		return "", fmt.Errorf("there was an error calling advapi32!LookupPrivilegeName for %+v with return code %d: %s", luid, ret, err)
	}

	var lpName uint8
	ret, _, err = lookupPrivilegeNameW.Call(0, uintptr(unsafe.Pointer(&luid)), uintptr(unsafe.Pointer(&lpName)), uintptr(unsafe.Pointer(&cchName)))
	if err != windows.Errno(0) || ret == 0 {
		return "", fmt.Errorf("there was an error calling advapi32!LookupPrivilegeName with return code %d: %s", ret, err)
	}

	return windows.BytePtrToString(&lpName), nil
}

type TOKEN_GROUPS struct {
	GroupCount uint32
	Groups     [1]windows.SIDAndAttributes
}

func GetTokenGroups(token windows.Token) (groups []windows.SIDAndAttributes, err error) {
	// Get the privileges and attributes
	// Call to get structure size
	var returnedLen uint32
	err = NtQueryInformationToken(token, windows.TokenGroups, nil, 0, &returnedLen)
	if returnedLen == 0 {
		err = fmt.Errorf("there was an error calling first windows.GetTokenInformation in GetTokenGroup: %s", err)
		return
	}

	// Call again to get the actual structure
	buf := make([]byte, returnedLen)
	err = NtQueryInformationToken(token, windows.TokenGroups, &buf[0], returnedLen, &returnedLen)
	if err != nil {
		err = fmt.Errorf("there was an error calling second windows.GetTokenInformation in GetTokenGroup: %s", err)
		return
	}

	tg := (*TOKEN_GROUPS)(unsafe.Pointer(&buf[0]))
	groupCount := tg.GroupCount
	Println("GroupCount is ", groupCount)

	groupsSlice := unsafe.Slice(&tg.Groups[0], groupCount)

	// Read in the LUID and Attributes
	for i := 1; i < int(groupCount); i++ {
		var group windows.SIDAndAttributes
		group = groupsSlice[i]
		Println(group.Sid.String())
		groups = append(groups, group)
	}
	return
}

func GetGroupNameFromSid(s *windows.SID) string {
	account, domain, _, err := s.LookupAccount("")
	if err == nil {
		return fmt.Sprintf("%s\\%s", domain, account)
	}
	return "Error in Lookup:" + err.Error()
}
