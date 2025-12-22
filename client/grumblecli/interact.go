package grumblecli

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/desertbit/grumble"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
)

func SetInteractCommands(conn grpc.ClientConnInterface) {
	interactCmd := &grumble.Command{
		Name: "interact",
		Help: "interact with a session",
        Run: func(c *grumble.Context) error {
			var sessionSelectedByUser string
			res, err := clientrpc.ListSessionFunc(conn)
            sessionList := res.GetSessions()
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			sessionSelector := &survey.Select{
				Message: "Specify a session :",
				Options: func() (out []string) {
					for _, session := range sessionList {
						if session.Status == "alive" {
							out = append(out, fmt.Sprintf("%s - %s@%s - %s", session.Id, session.User, session.Hostname, session.Ip))
						}
					}
					return
				}(),
			}
			err = survey.AskOne(sessionSelector, &sessionSelectedByUser)
			if err != nil {
				return err
			}

			s := strings.Split(sessionSelectedByUser, " ")
			sessionID := s[0]
			if err != nil {
				return err
			}


			for _, session := range sessionList {
                if session.Id == sessionID {
                    SelectedSession = session
                }
            }

            // Avoid having Hostname in Username if user is local
            user := SelectedSession.User
            temp := strings.SplitN(user, "\\", 2) 
            if len(temp) > 1 {
                if temp[0] == SelectedSession.Hostname {
                    user = temp[1]
                }
            }
            // Change Prompt
			c.App.SetPrompt(fmt.Sprintf("[Session %s: %s@%s] » ", SelectedSession.Id, user, SelectedSession.Hostname))

            // Reset Commands 
            SetCommands(conn)

			return nil
		},
	}
	app.AddCommand(interactCmd)
}
