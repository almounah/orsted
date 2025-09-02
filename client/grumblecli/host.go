package grumblecli

import (
	"fmt"
	"os"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetHosterCommands(conn grpc.ClientConnInterface) {
	hosterCmd := &grumble.Command{
		Name: "hoster",
		Help: "commands related to hosting and unhosting files",
	}

	hostCmd := &grumble.Command{
		Name: "host",
		Help: "command to host a file - currently on all listeners",
		Args: func(a *grumble.Args) {
			a.String("localfile", "Local File to Host on all listener")
			a.String("filename", "File name in the endpoint")
		},
		Run: func(c *grumble.Context) error {
			filePath := c.Args.String("localfile")
			b, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			err = clientrpc.HostFileFunc(conn, c.Args.String("filename"), b)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println("Hosted file for all listeners. Run hoster view.")
			return nil
		},
	}

	unhostCmd := &grumble.Command{
		Name: "unhost",
		Help: "command to unhost a file - currently on all listeners",
		Args: func(a *grumble.Args) {
			a.String("filename", "Only the file name, not full path")
		},
		Run: func(c *grumble.Context) error {
			err := clientrpc.UnHostFileFunc(conn, c.Args.String("filename"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println("Unhosted file")
			return nil
		},
	}

	viewhostCmd := &grumble.Command{
		Name: "view",
		Help: "command to view all hosted file - each is hosted on all listeners",
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.ViewHostFileFunc(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			for i := 0; i < len(res.GetHostlist()); i++ {
				data = append(data, []string{
					string(res.GetHostlist()[i].Filename),
				})
			}
			prettyPrint(data, []string{"Hosted file Endpoints"}, c.App.Stdout())
			return nil
		},
	}

	hosterCmd.AddCommand(hostCmd)
	hosterCmd.AddCommand(viewhostCmd)
	hosterCmd.AddCommand(unhostCmd)
	app.AddCommand(hosterCmd)
}
