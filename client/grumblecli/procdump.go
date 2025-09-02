package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetProcDumpCommand(conn grpc.ClientConnInterface) {
    downloadCmd := &grumble.Command{
		Name: "procdump",
		Help: "dump a process to a local file in the server",
		Args: func(a *grumble.Args) {
			a.String("pid", "pid of the process")
			a.String("path", "local destination on the server")
		},
		Flags: func (f *grumble.Flags)  {
			f.Bool("n", "native", false, "Dump process with native indirect syscall ntapi")
			
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			pid := c.Args.String("pid")
			path := c.Args.String("path")
			native := c.Flags.Bool("native")

			var nativeString string = ""
			var command string
			var prettycommand string

			if native {
				nativeString = "native"
			}
			command = fmt.Sprintf("procdump %s %s %s", pid, path, nativeString)
			prettycommand = command

			if native {
				prettycommand = fmt.Sprintf("procdump --%s %s %s", nativeString, pid, path)

			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, []byte{}, prettycommand)
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
	app.AddCommand(downloadCmd)
}
