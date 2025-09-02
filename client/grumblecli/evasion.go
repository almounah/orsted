package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetEvasionCommand(conn grpc.ClientConnInterface) {
	evasionCmd := &grumble.Command{
		Name: "evasion",
		Help: "Evasion AMSI/ETW and unhook ntdll with indirect syscall",
	}

    amsiCmd := &grumble.Command{
		Name: "amsi",
		Help: "Evade AMSI using indirect syscall",
		Args: func(f *grumble.Args) {
            s := "The number of the method used to patch AMSI\n"
            s += " - 1 - Will patch beginning of AmsiScanBuffer in amsi.dll with indirect syscall \n"
            s += " - 2 - Will patch a je to jne of AmsiScanBuffer in amsi.dll with indirect syscall \n"
            s += " - 3 - Will patch AmsiOpenSession Context in amsi.dll with indirect syscall \n"
            s += " - 4 - Will patch beginning of AmsiOpenSession in amsi.dll with indirect syscall \n"
            s += " - 5 - Will patch a je to jne of AmsiOpenSession in amsi.dll with indirect syscall \n"
            s += " - 6 - Will patch a je to jne of AmsiScanString in amsi.dll with indirect syscall \n"
            s += " - 7 - Will patch Amsi by tampering clr.dll with indirect syscall \n"
            s += " - 8 - Will patch AmsiScanBuffer by placing HardwareBreakpoint \n"
			f.String("method", s)
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "evasion amsi " + c.Args.String("method"), nil, "evasion amsi " + c.Args.String("method"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			data = append(data, []string{
				res.TaskId,
				res.BeacondId,
				res.State,
				string(res.Command),
			})
			prettyPrint(data, []string{"TASKID", "SESSIONID", "STATE", "COMMAND"}, c.App.Stdout())
			return nil
		},
	}

    etwCmd := &grumble.Command{
		Name: "etw",
		Help: "Evade ETW using indirect syscall",
		Args: func(f *grumble.Args) {
            s := "The number of the method used to patch ETW\n"
            s += " - 1 - Will patch EtwpEventWriteFull in ntdll.dll with indirect syscall \n"
            s += " - 2 - Will patch EtwEventWrite, EtwEventWriteFull EtwEventWriteEx in ntdll with indirect syscall \n"
            s += " - 3 - Will patch EtwEventWrite, EtwEventWriteEx in advapi32.dll with indirect syscall \n"
            s += " - 4 - Will patch NtTraceEvent in ntdll.dll with indirect syscall \n"
            s += " - 5 - Will patch EtwpEventWriteFull by placing HardwareBreakpoint \n"
			f.String("method", s)
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "evasion etw " + c.Args.String("method"), nil, "evasion etw " + c.Args.String("method"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			data = append(data, []string{
				res.TaskId,
				res.BeacondId,
				res.State,
				string(res.Command),
			})
			prettyPrint(data, []string{"TASKID", "SESSIONID", "STATE", "COMMAND"}, c.App.Stdout())
			return nil
		},
	}

    evasionCmd.AddCommand(etwCmd)
    evasionCmd.AddCommand(amsiCmd)
	app.AddCommand(evasionCmd)
}

