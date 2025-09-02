package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetlsCommand(conn grpc.ClientConnInterface) {
	lsCommandCmd := &grumble.Command{
		Name: "ls",
		Help: "List file in a directory",
		Args: func(f *grumble.Args) {
			f.StringList("path", "path to list")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
            pathString := ""
            pathStringList := c.Args.StringList("path")
            for i := 0; i < len(pathStringList); i++ {
                pathString += pathStringList[i]
				pathString += " "
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "ls "+pathString, []byte{}, "ls "+pathString)
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
	app.AddCommand(lsCommandCmd)
}
