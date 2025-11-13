package grumblecli

import (
	"encoding/json"
	"fmt"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

type WinrmAuthparam struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Hash     string `json:"hash"`
	Command  string `json:"command"`
}

func SetWinrmCommand(conn grpc.ClientConnInterface) {
	psexecCmd := &grumble.Command{
		Name: "winrm",
		Help: "winrm on a remote host",
		Flags: func(f *grumble.Flags) {
			f.String("i", "host", "localhost", "The remote host")
			f.String("P", "port", "5895", "The remote winrm port")
			f.Bool("", "insecure", false, "Insecure connection")
			f.Bool("", "tls", false, "Use HTTPS")
			f.Bool("", "background", false, "Don't catch output, just run in background")
			f.String("u", "username", "rudeus", "The username to run winrm as ex. rudeus, Corp\\\\Rudeus, rudeus@corp")
			f.String("p", "password", "", "The password to run winrm as")
			f.String("H", "hash", "", "The password to run winrm as")
			f.String("c", "command", "whoami", "The command to run")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			//TODO Add space check
			var host string = c.Flags.String("host")
			var port string = c.Flags.String("port")
			var insecure string = map[bool]string{true: "insecure", false: "not-ionsecure"}[c.Flags.Bool("insecure")]
			var tls string = map[bool]string{true: "tls", false: "not-tls"}[c.Flags.Bool("tls")]
			var background string = map[bool]string{true: "background", false: "not-background"}[c.Flags.Bool("background")]
			var username string = c.Flags.String("username")
			var password string = c.Flags.String("password")
			var hash string = c.Flags.String("hash")
			var commandExec string = c.Flags.String("command")

			var authType string
			var winrmAuthparam WinrmAuthparam
			winrmAuthparam.Username = username
			winrmAuthparam.Password = password
			winrmAuthparam.Hash = hash
			winrmAuthparam.Command = commandExec
			b, err := json.Marshal(winrmAuthparam)
			if err != nil {
				fmt.Println(err.Error())
				return nil
			}

			// Username is empty mean we require kerberos auth
			if username == "" {
				authType = "kerberos"
				// Domain is empty it means it is a local auth to remote host
			} else {
				authType = "ntlm"
			}

			command := fmt.Sprintf("winrm %s %s %s %s %s %s", host, port, insecure, tls, authType, background)
			prettycommand := fmt.Sprintf("winrm --host %s --port %s --username %s --password %s --hash %s --command \"%s\"", host, port, username, password, hash, commandExec)
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, command, b, prettycommand)
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

	app.AddCommand(psexecCmd)
}
