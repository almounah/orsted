package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetSleepCommand(conn grpc.ClientConnInterface) {
	sleepCmd := &grumble.Command{
		Name: "sleep",
		Help: "Change Sleep of beacon",
		Args: func(a *grumble.Args) {
			a.String("interval", "sleep in millisecond")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "sleep "+c.Args.String("interval"), nil, "sleep "+c.Args.String("interval"))
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
	app.AddCommand(sleepCmd)
}
