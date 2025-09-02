package grumblecli

import (

	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

)

func SetBackCommands(conn grpc.ClientConnInterface) {
	interactCmd := &grumble.Command{
		Name: "back",
		Help: "back to no selected session",
        Run: func(c *grumble.Context) error {
            // Change Prompt
			c.App.SetPrompt("orsted-client » ")

            // Change Selected session
            SelectedSession = nil

            // Reset Commands 
            SetCommands(conn)

			return nil
		},
	}
	app.AddCommand(interactCmd)
}
