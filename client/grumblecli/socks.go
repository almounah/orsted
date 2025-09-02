package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetSocksCommand(conn grpc.ClientConnInterface) {
	socksCmd := &grumble.Command{
		Name: "socks",
		Help: "Commands related to the socks",
	}

	startCmd := &grumble.Command{
		Name: "start",
		Help: "Start the socks proxy locally",
		Flags: func(f *grumble.Flags) {
			f.String("s", "session", "", "Session to Bind socks to")
			f.String("i", "ip", "127.0.0.1", "Ip address for socks - 127.0.0.1 default")
			f.String("p", "port", "1080", "port for socks")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			_, err := clientrpc.StartSocks(conn, c.Flags.String("session"), c.Flags.String("ip"), c.Flags.String("port"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println("Created Socks with beacon")
			return nil
		},
	}

	bindCmd := &grumble.Command{
		Name: "bind",
		Help: "bind socks to selected beacon",
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "socks bind", []byte{}, "socks bind")
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println(res)
			return nil
		},
	}

	unbindCmd := &grumble.Command{
		Name: "unbind",
		Help: "unbind socks from selected beacon",
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "socks unbind", []byte{}, "socks unbind")
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println(res)
			return nil
		},
	}

	socksCmd.AddCommand(bindCmd)
	socksCmd.AddCommand(unbindCmd)
	socksCmd.AddCommand(startCmd)
	app.AddCommand(socksCmd)
}
