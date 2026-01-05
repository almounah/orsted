package grumblecli

import (
	"fmt"
	"strings"

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)
func SetListenerCommands(conn grpc.ClientConnInterface) {
	listenerCmd := &grumble.Command{
		Name: "listener",
		Help: "Commands related to the listener",
	}

	startCmd := &grumble.Command{
		Name: "start",
		Help: "Start the listener",
		Args: func(f *grumble.Args) {
            f.String("type", "type of the listener: http or https")
			f.String("ip", "ip of the listener")
			f.String("port", "port")
		},
		Completer: func(prefix string, args []string) []string {
			listenertypeList := []string{"http", "https"}
			ipList := []string{"127.0.0.1", "0.0.0.0"}
			portList := []string{"80", "443", "4444"}
			var suggestions []string

            var modulesList []string
            if len(args) == 0 {
                modulesList = listenertypeList
            }
            if len(args) == 1 {
                modulesList = ipList
            }
            if len(args) == 2 {
                modulesList = portList
            }

			for _, moduleName := range modulesList {
				if strings.HasPrefix(moduleName, prefix) {
					suggestions = append(suggestions, moduleName)
				}
			}
			return suggestions
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
            listenerType := c.Args.String("type")
            listenerIp := c.Args.String("ip")
            listenerport := c.Args.String("port")


			var data [][]string
            switch listenerType {
            case "http":
                res, err := clientrpc.StartHttpListenerFunc(conn, listenerIp, listenerport)
                if err != nil {
                    fmt.Println("Error Occured ", err.Error())
                    return nil
                }
                data = [][]string{
                    {res.Id, res.Ip, res.Port, res.ListenerType},
                }
            case "https":
                certPath := Conf.DefaultHTTPSCert
                keyPath := Conf.DefaultHTTPSKey
                fmt.Println("Using the following certificate and key on server: ", certPath, keyPath)
                res, err := clientrpc.StartHttpsListenerFunc(conn, listenerIp, listenerport, certPath, keyPath)
                if err != nil {
                    fmt.Println("Error Occured ", err.Error())
                    return nil
                }
                data = [][]string{
                    {res.Id, res.Ip, res.Port, res.ListenerType},
                }
            }

            // Print Some output
			c.App.Println("Created Listener")
			prettyPrint(data, []string{"ID", "IP", "PORT", "LISTENER TYPE"}, c.App.Stdout())
			return nil
		},
	}

	listCmd := &grumble.Command{
		Name: "list",
		Help: "list current listener",
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			res, err := clientrpc.ListListenerFunc(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			for i := 0; i < len(res.GetListener()); i++ {
				data = append(data, []string{res.GetListener()[i].Id, 
					res.GetListener()[i].ListenerType,
					res.GetListener()[i].Ip,
					res.GetListener()[i].Port})
			}
			prettyPrint(data, []string{"ID", "TYPE", "IP", "PORT"}, c.App.Stdout())
			return nil
		},
	}

	deleteCmd := &grumble.Command{
		Name: "stop",
		Help: "Stop and Delete a listener",
		Args: func(f *grumble.Args) {
			f.String("id", "Id of the listener")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			err := clientrpc.DeleteListenerFunc(conn, c.Args.String("id"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			c.App.Println("Deleted Listener ", c.Args.String("id"))
			return nil
		},
	}

	listenerCmd.AddCommand(startCmd)
	listenerCmd.AddCommand(listCmd)
	listenerCmd.AddCommand(deleteCmd)
	app.AddCommand(listenerCmd)
}
