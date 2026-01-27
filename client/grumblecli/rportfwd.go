package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetRportFwdCommand(conn grpc.ClientConnInterface) {
	autoRouteCmd := &grumble.Command{
		Name: "rportfwd",
		Help: "Command related to ligolo rportfwd. It works closely with autoroute command",
	}
	addCmd := &grumble.Command{
		Name: "add",
		Help: "Add rportfwd",
		Flags: func(f *grumble.Flags) {
			f.String("l", "local", "", "Local Address to Port Forward. Ex 127.0.0.1:4444")
			f.String("r", "remote", "", "Remote Address to Port Forward. Ex. 0.0.0.0:8000")
		},
		Run: func(c *grumble.Context) error {
			localAddress := c.Flags.String("local")
			remoteAddress := c.Flags.String("remote")
			if localAddress == "" || remoteAddress == "" {
				fmt.Println("Cannot Provide empty addresses. Use --local --remote to provide addresses.")
				return nil
			}
			_, err := clientrpc.AddRouteForRevPortFwd(conn, SelectedSession.Id, remoteAddress, localAddress)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println("Reverse Port Forward Added Successfully")
			return nil
		},
	}
	listCmd := &grumble.Command{
		Name: "list",
		Help: "list rportfwd",
		Run: func(c *grumble.Context) error {
			res, err := clientrpc.ListRoute(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			for i := 0; i < len(res.Routes); i++ {
				data = append(data, []string{res.Routes[i].RouteId, res.Routes[i].BeaconId, res.Routes[i].Subnet, res.Routes[i].Rportfwd})
			}
			prettyPrint(data, []string{"ROUTE ID", "BEACON ID", "SUBNET", "RPORTFWD (Local <-> Remote)"}, c.App.Stdout())
			return nil
		},
	}
	deleteCmd := &grumble.Command{
		Name: "delete",
		Help: "delete route subnet. If route subnet becomes empty, delete route on the fly.",
		Args: func(a *grumble.Args) {
			a.String("beaconId", "Beacon affected by route")
			a.String("remoteSrc", "subnet to be routed through beacon")
		},
		Run: func(c *grumble.Context) error {
			_, err := clientrpc.DeleteRevPortFwd(conn, c.Args.String("beaconId"), c.Args.String("remoteSrc"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println("Rportfwd Deleted Successfully. If Subnet and RportForward empty route will be deleted. You need to wait about 1min to be able to re-autoroute via same beacon (needed for graceful close)")
			return nil
		},
	}
	autoRouteCmd.AddCommand(deleteCmd)
	autoRouteCmd.AddCommand(addCmd)
	autoRouteCmd.AddCommand(listCmd)
	app.AddCommand(autoRouteCmd)
}
