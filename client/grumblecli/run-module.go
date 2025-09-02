package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetRunCommand(conn grpc.ClientConnInterface) {
	runCommandCmd := &grumble.Command{
		Name: "run",
		Help: "run a command",
		Args: func(f *grumble.Args) {
			f.StringList("command", "command to run")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
            commandString := ""
            commandStringList := c.Args.StringList("command")
            for i := 0; i < len(commandStringList); i++ {
                commandString += commandStringList[i]
				commandString += " "
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "run "+commandString, []byte{}, "run "+commandString)
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
	app.AddCommand(runCommandCmd)
}
