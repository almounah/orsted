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
	Domain   string `json:"domain"`
	Command  string `json:"command"`
}

func SetWinrmCommand(conn grpc.ClientConnInterface) {
	psexecCmd := &grumble.Command{
		Name: "winrm",
		Help: "winrm on a remote host",
		Flags: func(f *grumble.Flags) {
			f.String("H", "host", "localhost", "The remote host")
			f.String("P", "port", "5895", "The remote winrm port")
			f.Bool("i", "insecure", false, "Insecure connection")
			f.Bool("t", "tls", false, "Use HTTPS")
			f.String("u", "username", "rudeus", "The username to run winrm as")
			f.String("p", "password", "rudeus", "The password to run winrm as")
			f.String("d", "domain", "", "The domain, keep empty for local auth")
			f.String("c", "command", "whoami /all", "The comamnd to run")
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
			var username string = c.Flags.String("username")
			var password string = c.Flags.String("password")
			var domain string = c.Flags.String("domain")
			var commandExec string = c.Flags.String("command")

			var authType string
			var winrmAuthparam WinrmAuthparam
			winrmAuthparam.Username = username
			winrmAuthparam.Password = password
			winrmAuthparam.Domain = domain
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
			} else if domain == "" {
				authType = "local"
				// Otherwise it is username password auth with domain user
			} else {
				authType = "domain"
			}

			command := fmt.Sprintf("winrm %s %s %s %s %s", host, port, insecure, tls, authType)
			prettycommand := fmt.Sprintf("winrm --host %s --port %s --username %s --password %s --command %s", host, port, username, password, commandExec)
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
