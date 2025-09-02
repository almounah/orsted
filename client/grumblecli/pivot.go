package grumblecli

import (
	"fmt"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetPivotCommand(conn grpc.ClientConnInterface) {
	pivotCmd := &grumble.Command{
		Name: "pivot",
		Help: "Commands related to the pivot",
	}

	startCmd := &grumble.Command{
		Name: "start",
		Help: "Start the pivot on a specific beacon",
		Completer: func(prefix string, args []string) []string {
			transportList := []string{"tcp", "smb"}
			addressList := []string{"127.0.0.1:4444", "namedpipe1"} 
			var suggestions []string

			var modulesList []string
			if len(args) == 0 {
				modulesList = transportList
			}
			if len(args) == 1 {
				modulesList = addressList
			}

			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName)
				}
			}
			return suggestions
		},
		Args: func(f *grumble.Args) {
            f.String("pivtype", "Type of the pivot: tcp, smb")
			f.String("address", "Address string: 127.0.0.1:4444, namedpipe1")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			pivType := c.Args.String("pivtype")
			address := c.Args.String("address")
			command := fmt.Sprintf("pivot start %s %s", pivType, address)
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, []byte{}, command)
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

	stopCmd := &grumble.Command{
		Name: "stop",
		Help: "Stop the pivot on a specific beacon",
		Args: func(f *grumble.Args) {
            f.String("pivId", "Id of the pivot")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			pivId := c.Args.String("pivId")
			command := fmt.Sprintf("pivot stop %s", pivId)
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, []byte{}, command)
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


	listCmd := &grumble.Command{
		Name: "list",
		Help: "List pivot on as beacon",
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			command := fmt.Sprintf("pivot list")
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, []byte{}, command)
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

    pivotCmd.AddCommand(startCmd)
    pivotCmd.AddCommand(stopCmd)
    pivotCmd.AddCommand(listCmd)
    app.AddCommand(pivotCmd)
}
