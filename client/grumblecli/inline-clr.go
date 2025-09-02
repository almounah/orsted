package grumblecli

import (
	"fmt"
	"os"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)


func SetInlineExecCommand(conn grpc.ClientConnInterface) {
	inlineExecuteCmd := &grumble.Command{
		Name: "inline-clr",
		Help: "Load and Execute NET Assembly in Memory",
	}

    loadClrCmd := &grumble.Command{
		Name: "start-clr",
		Help: "Load CLR v4",
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "inline-clr start-clr", nil, "inline-clr start-clr")
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

    loadAssemblyCmd := &grumble.Command{
		Name: "load-assembly",
		Help: "Load Assmebly in CLR",
		Args: func(f *grumble.Args) {
			f.String("file", "Assembly to load")
		},
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
			moduleName := c.Args.String("file")
			b, err := os.ReadFile(Conf.NetAssemblyPath + moduleName)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
                return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "inline-clr load-assembly "+c.Args.String("file"), b, "inline-clr load-assembly "+c.Args.String("file"))
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

    invokeAssemblyCmd := &grumble.Command{
		Name: "invoke-assembly",
		Help: "Invoke Assembly in CLR",
		Args: func(f *grumble.Args) {
			f.String("file", "Assembly to load")
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
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "inline-clr invoke-assembly "+c.Args.String("file") + " " + argString, nil, "inline-clr invoke-assembly "+c.Args.String("file") + " " + argString)
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

    listAssemblyCmd := &grumble.Command{
		Name: "list-assemblies",
		Help: "List Assembly loaded in CLR",
		Run: func(c *grumble.Context) error {
            if SelectedSession == nil {
                fmt.Println("No session selected. Use interact command to specify session")
                return nil
            }
            res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "inline-clr list-assemblies", nil, "incline-clr list-assemblies")
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

    inlineExecuteCmd.AddCommand(listAssemblyCmd)
    inlineExecuteCmd.AddCommand(loadClrCmd)
    inlineExecuteCmd.AddCommand(loadAssemblyCmd)
    inlineExecuteCmd.AddCommand(invokeAssemblyCmd)
	app.AddCommand(inlineExecuteCmd)
}

