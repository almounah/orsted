package grumblecli

import (
	"fmt"
	"os"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetPowercliffCommand(conn grpc.ClientConnInterface) {
	pwCliffCmd := &grumble.Command{
		Name: "powercliff",
		Help: "Powercliff, inline powershell with superdeye managed patched",
	}

	startPwCliffCmd := &grumble.Command{
		Name: "start-powercliff",
		Help: "Load CLR, instantiate powershell and patch some stuff",
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "powercliff start-powercliff", nil, "powercliff start-powercliff")
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

	loadPs1Cmd := &grumble.Command{
		Name: "load",
		Help: "Load ps1 script to powershell",
		Args: func(f *grumble.Args) {
			f.String("file", "Ps1 to load")
		},
		Completer: func(prefix string, args []string) []string {
			batcaveSuggestion := GetListOfGadgetName("ps1") 
			var suggestions []string

            var modulesList []string
            if len(args) == 0 {
                modulesList = batcaveSuggestion
            }
			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName + ".ps1")
				}
			}
			return suggestions
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			moduleName := c.Args.String("file")
			DownloadAndUnzipBatGadget(strings.TrimSuffix(moduleName, ".ps1"), false)
			b, err := os.ReadFile(Conf.Ps1ScriptPath + moduleName)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "powercliff exec", b, "powercliff load "+c.Args.String("file"))
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

	execCmd := &grumble.Command{
		Name: "exec",
		Help: "Exec a powershell cmd",
		Args: func(f *grumble.Args) {
			f.StringList("args", "Argument of the Assembly")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			argString := ""
			argStringList := c.Args.StringList("args")
			for i := 0; i < len(argStringList); i++ {
				argString += argStringList[i]
				argString += " "
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "powercliff exec", []byte(argString), "powercliff exec " + argString)
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


	pwCliffCmd.AddCommand(startPwCliffCmd)
	pwCliffCmd.AddCommand(loadPs1Cmd)
	pwCliffCmd.AddCommand(execCmd)
	app.AddCommand(pwCliffCmd)
}
