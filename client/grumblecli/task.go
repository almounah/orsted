package grumblecli

import (
	"fmt"
	"os"
	"strings"

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
		Args: func(f *grumble.Args) {
			f.String("sessionid", "Id of the session")
		},
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.ListTaskFunc(conn, c.Args.String("sessionid"))
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
		Args: func(f *grumble.Args) {
			f.String("taskid", "Id Of the task")
		},
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.GetSingleTaskFunc(conn, c.Args.String("taskid"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println(fmt.Sprintf("<< %s \n>> %s", res.GetPrettyCommand(), res.GetResponse()))
			return nil
		},
	}

	exportCmd := &grumble.Command{
		Name: "export",
		Help: "export the task result to a file",
		Args: func(f *grumble.Args) {
			f.String("taskid", "Id Of the task")
			f.String("file", "file to export too (on the client)")
		},
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.GetSingleTaskFunc(conn, c.Args.String("taskid"))
			file := c.Args.String("file")
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			if file != "" {
				if err := os.WriteFile(file, []byte(res.GetResponse()), 0644); err != nil {
					fmt.Println("Failed to write file:", err)
				} else {
					fmt.Println("Response written to", file)
				}
			}
			return nil
		},
	}

	searchCmd := &grumble.Command{
		Name: "search",
		Help: "search for task among all tasks",
		Args: func(f *grumble.Args) {
			f.String("param", "String to search for")
		},
		Run: func(c *grumble.Context) error {
			searchParam := c.Args.String("param")

			res, err := clientrpc.ListSessionFunc(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string

			for i := 0; i < len(res.GetSessions()); i++ {
				id := res.GetSessions()[i].Id
				res, err := clientrpc.ListTaskFunc(conn, id)
				if err != nil {
					fmt.Println("Error Occured ", err.Error())
					return nil
				}
				for j := 0; j < len(res.GetTasks()); j++ {
					if strings.Contains(res.GetTasks()[j].PrettyCommand, searchParam) {
					data = append(data, []string{
						res.GetTasks()[j].TaskId,
						res.GetTasks()[j].BeacondId,
						res.GetTasks()[j].State,
						string(res.GetTasks()[j].PrettyCommand),
					})
					}
				}
			}
			prettyPrint(data, []string{"TASKID", "SESSIONID", "STATE", "COMMAND"}, c.App.Stdout())
			return nil
		},
	}

	taskCmd.AddCommand(listCmd)
	taskCmd.AddCommand(searchCmd)
	taskCmd.AddCommand(exportCmd)
	taskCmd.AddCommand(viewCmd)
	app.AddCommand(taskCmd)
}
