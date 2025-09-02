package grumblecli

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetShellCommands(conn grpc.ClientConnInterface) {
	shellCmd := &grumble.Command{
		Name: "shell",
		Help: "start and interact with shell (not opsec)",
	}

	startCmd := &grumble.Command{
		Name: "start",
		Help: "start interactive shell",
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "shell start", []byte{}, "shell start")
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

	listCmd := &grumble.Command{
		Name: "list",
		Help: "list all interactive shell",
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "shell list", []byte{}, "shell list")
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

	interactStartCmd := &grumble.Command{
		Name: "interact",
		Help: "interact with interactive shell",
		Args: func(f *grumble.Args) {
			f.String("pid", "Pid of the shell")
		},
		Run: func(c *grumble.Context) error {
			if SelectedSession == nil {
				fmt.Println("No session selected. Use interact command to specify session")
				return nil
			}
			res, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "shell interact-start "+c.Args.String("pid"), []byte{}, "shell interact-start "+c.Args.String("pid"))
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
			fmt.Println("< Will Interact with Shell >")
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			// Done channel to break the loop
			done := make(chan struct{})

			// Start a goroutine to listen for Ctrl+C
			go func() {
				<-sigChan
				fmt.Println("\nReceived interrupt, exiting loop...")
				close(done)
			}()

			for {
				select {
				case <-done:
					// Exit the loop
					fmt.Println("Exiting Shell")
					res, err = clientrpc.AddTaskFunc(conn, SelectedSession.Id, "shell interact-stop "+c.Args.String("pid"), []byte{}, "shell interact-stop "+c.Args.String("pid"))
					if err != nil {
						fmt.Println(err.Error())
						return nil
					}
					return nil
				default:
					var input string
					// fmt.Print("Pseudo shell > ")
					scanner := bufio.NewScanner(os.Stdin)
					scanner.Scan()
					input = scanner.Text()
					_, err := clientrpc.AddTaskFunc(conn, SelectedSession.Id, "shell interact-write "+c.Args.String("pid")+" "+input, []byte{}, "shell interact-write "+c.Args.String("pid")+" "+input)
					if err != nil {
						fmt.Println("Error Occured ", err.Error())
					}
				}
			}
		},
	}
	shellCmd.AddCommand(startCmd)
	shellCmd.AddCommand(listCmd)
	shellCmd.AddCommand(interactStartCmd)
	app.AddCommand(shellCmd)
}
