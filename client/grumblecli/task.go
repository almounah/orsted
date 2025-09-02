package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetGeneralTaskCommands(conn grpc.ClientConnInterface) {
	taskCmd := &grumble.Command{
		Name: "task",
		Help: "Commands related to the beacon tasks",
	}

	listCmd := &grumble.Command{
		Name: "list",
		Help: "list current sessions",
		Flags: func(f *grumble.Flags) {
			f.String("s", "session", "", "Id of the session")
		},
		Run: func(c *grumble.Context) error {
			if c.Flags.String("session") == "" {
				fmt.Println("Please select a session with -s flag")
				return nil 
			}
			res, err := clientrpc.ListTaskFunc(conn, c.Flags.String("session"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			for i := 0; i < len(res.GetTasks()); i++ {
				data = append(data, []string{
					res.GetTasks()[i].TaskId,
					res.GetTasks()[i].BeacondId,
					res.GetTasks()[i].State,
					string(res.GetTasks()[i].PrettyCommand),
				})
			}
			prettyPrint(data, []string{"TASKID", "SESSIONID", "STATE", "COMMAND"}, c.App.Stdout())
			return nil
		},
	}

	viewCmd := &grumble.Command{
		Name: "view",
		Help: "print the output of a session",
		Flags: func(f *grumble.Flags) {
			f.String("t", "taskid", "", "Id Of the task")
		},
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.GetSingleTaskFunc(conn, c.Flags.String("taskid"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println(fmt.Sprintf("<< %s \n>> %s", res.GetPrettyCommand(), res.GetResponse()))
			return nil
		},
	}
	taskCmd.AddCommand(listCmd)
	taskCmd.AddCommand(viewCmd)
	app.AddCommand(taskCmd)
}
