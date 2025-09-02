package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)


func SetRunAsCommand(conn grpc.ClientConnInterface) {
	runasCmd := &grumble.Command{
		Name: "runas",
		Help: "RunAs a process in go",
		Flags: func(f *grumble.Flags) {
			f.Bool("b", "no-background", false, "Specify to not run process in background and capture output")
		},
		Args: func(f *grumble.Args) {
			f.String("username", "Username ex. rudeus, CORP\\rudeus")
			f.String("password", "Password of username")
			f.StringList("app", "Application to run with its arguments")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
            argString := ""
            argStringList := c.Args.StringList("app")
            for i := 0; i < len(argStringList); i++ {
                argString += argStringList[i]
				argString += " "
            }

			background := "background"
			if c.Flags.Bool("no-background") {
				background = "no-background"
			}

			commandToRun := fmt.Sprintf("runas %s %s %s %s", background, c.Args.String("username"), c.Args.String("password"), argString)
			prettycommand := fmt.Sprintf("runas --%s %s %s %s", background, c.Args.String("username"), c.Args.String("password"), argString)
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, commandToRun, nil, prettycommand)
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


	app.AddCommand(runasCmd)
}

