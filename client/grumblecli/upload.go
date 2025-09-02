package grumblecli

import (
	"fmt"
	"os"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetUploadCommands(conn grpc.ClientConnInterface) {
    uploadCmd := &grumble.Command{
		Name: "upload",
		Help: "upload file from server to beacon",
		Args: func(a *grumble.Args) {
			a.String("filetoupload", "file to upload")
			a.String("remotepath", "remote path")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
            uploaddata, err := os.ReadFile(c.Args.String("filetoupload"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "upload "+c.Args.String("remotepath"), uploaddata, "upload "+c.Args.String("remotepath"))
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
	app.AddCommand(uploadCmd)
}
