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
		Help: "Add revportfwd",
		Args: func(a *grumble.Args) {
			a.String("beaconId", "Beacon affected by route")
			a.String("remoteSrc", "Remote Address")
			a.String("localDst", "Local Address")
		},
		Run: func(c *grumble.Context) error {
			_, err := clientrpc.AddRouteForRevPortFwd(conn, c.Args.String("beaconId"), c.Args.String("remoteSrc"), c.Args.String("localDst"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println("Route Added successfully")
			return nil
		},
	}
	listCmd := &grumble.Command{
		Name: "list",
		Help: "list route",
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
			prettyPrint(data, []string{"ROUTE ID", "BEACON ID", "SUBNET", "REVPORTFWD"}, c.App.Stdout())
			return nil
		},
	}
	deleteCmd := &grumble.Command{
		Name: "delete",
		Help: "delete route subnet. If route subnet becomes empty, delete route o the fly.",
		Args: func(a *grumble.Args) {
			a.String("beaconId", "Beacon affected by route")
			a.String("subnet", "subnet to be routed through beacon")
		},
		Run: func(c *grumble.Context) error {
			_, err := clientrpc.DeleteRoute(conn, c.Args.String("beaconId"), c.Args.String("subnet"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println("Subnet Deleted Successfully. If Subnet empty route will be deleted. You need to wait about 1min to be able to re-autoroute via same beacon (needed for graceful close)")
			return nil
		},
	}
	autoRouteCmd.AddCommand(deleteCmd)
	autoRouteCmd.AddCommand(addCmd)
	autoRouteCmd.AddCommand(listCmd)
	app.AddCommand(autoRouteCmd)
}
