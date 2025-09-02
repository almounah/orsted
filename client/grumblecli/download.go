package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetDownloadCommands(conn grpc.ClientConnInterface) {
    downloadCmd := &grumble.Command{
		Name: "download",
		Help: "download file from beacon to the server",
		Args: func(a *grumble.Args) {
			a.String("remotefile", "remote file to download")
			a.String("localpath", "local destination on the server")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "download "+c.Args.String("remotefile")+ " " + c.Args.String("localpath"), []byte{}, "download "+c.Args.String("remotefile")+ " " + c.Args.String("localpath"))
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
