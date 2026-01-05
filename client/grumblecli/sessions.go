package grumblecli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/desertbit/grumble"
	"github.com/xlab/treeprint"
	"google.golang.org/grpc"

	"orsted/client/clientrpc"
	"orsted/protobuf/orstedrpc"
)

func getBeaconString(beaconId string, sesstionList []*orstedrpc.Session) string {
	for i := 0; i < len(sesstionList); i++ {
		if beaconId == sesstionList[i].Id {
			session := sesstionList[i]
			user := session.User
			// TO not get hostname in user
			temp := strings.SplitN(user, "\\", 2)
			if len(temp) > 1 {
				if temp[0] == session.Hostname {
					user = temp[1]
				}
			}

			res := fmt.Sprintf("(%s/%s) %s: %s@%s", session.Transport, session.Os, beaconId, user, session.Ip)
			return res
		}
	}
	return "ERROR in GETTING INFO"
}

func addBeaconToTree(beaconId string, chain []string, addedBeaconToTree map[string]treeprint.Tree, sessionList []*orstedrpc.Session) {
	// Already added
	if _, exists := addedBeaconToTree[beaconId]; exists {
		return
	}

	// Determine parent
	var parentId string
	if len(chain) > 1 {
		parentId = chain[1]
	} else {
		parentId = "0" // root
	}

	// Parent exists in tree
	if parentTree, exists := addedBeaconToTree[parentId]; exists {
		addedBeaconToTree[beaconId] = parentTree.AddBranch(getBeaconString(beaconId, sessionList))
	} else {
		// Parent not added yet, recurse up the chain
		if len(chain) > 1 {
			addBeaconToTree(parentId, chain[1:], addedBeaconToTree, sessionList)
		} else {
			// Add root if not present
			addedBeaconToTree[parentId] = treeprint.New()
		}
		// Retry after parent has been added
		addBeaconToTree(beaconId, chain, addedBeaconToTree, sessionList)
	}
}

func SetSessionCommands(conn grpc.ClientConnInterface) {
	sessionCmd := &grumble.Command{
		Name: "session",
		Help: "Commands related to the beacon sessions",
	}

	listCmd := &grumble.Command{
		Name: "list",
		Help: "list current sessions",
		Flags: func(f *grumble.Flags) {
			f.Bool("a", "all", false, "if specified will show all session. even stopped one")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener

			all := c.Flags.Bool("all")
			res, err := clientrpc.ListSessionFunc(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var data [][]string
			for i := 0; i < len(res.GetSessions()); i++ {
				if all || res.GetSessions()[i].Status == "alive" {
					data = append(data, []string{
						res.GetSessions()[i].Id,
						res.GetSessions()[i].Ip,
						res.GetSessions()[i].Hostname,
						res.GetSessions()[i].User,
						res.GetSessions()[i].Integrity,
						res.GetSessions()[i].Os,
						strconv.FormatInt(time.Now().Unix()-res.GetSessions()[i].Lastseen, 10),
						res.GetSessions()[i].Status,
					})
				}
			}
			prettyPrint(data, []string{"ID", "IP", "HOSTNAME", "USER", "INTEGRITY", "OS", "POL", "STATUS"}, c.App.Stdout())
			return nil
		},
	}

	treeCmd := &grumble.Command{
		Name: "tree",
		Help: "tree print the sessions",
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			res, err := clientrpc.ListSessionFunc(conn)
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			var addedSessionInTree map[string]treeprint.Tree
			addedSessionInTree = make(map[string]treeprint.Tree)
			tree := treeprint.New()
			treeprint.IndentSize = 7
			// Added Server / Firewall Node
			addedSessionInTree["0"] = tree
			for i := 0; i < len(res.GetSessions()); i++ {
				chain := res.GetSessions()[i].GetChain()
				beaconId := res.GetSessions()[i].Id
				// Each chain has "0" as a parent
				// Maybe move this to orsteddb ?
				// TODO
				chain = append(chain, "0")
				chain = append([]string{beaconId}, chain...)
				for j := 0; j < len(chain); j++ {
					addBeaconToTree(chain[j], chain, addedSessionInTree, res.GetSessions())
				}

			}

			fmt.Println(addedSessionInTree["0"].String())
			return nil
		},
	}

	stopCmd := &grumble.Command{
		Name: "stop",
		Help: "stop the session by sending stop task to beacon and marking beacon as stopped",
		Args: func(f *grumble.Args) {
			f.String("id", "id of the session")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			err := clientrpc.StopSessionByIdFunc(conn, c.Args.String("id"))
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
			}
			fmt.Println("Stopped session", c.Args.String("id"))
			return nil
		},
	}

	interactCmd := &grumble.Command{
		Name: "interact",
		Help: "provide another way to interact with session",
		Args: func(f *grumble.Args) {
			f.String("id", "id of the session")
		},
		Run: func(c *grumble.Context) error {
			// Implement the logic to start the listener
			sessionID := c.Args.String("id")
			res, err := clientrpc.ListSessionFunc(conn)
			sessionList := res.GetSessions()
			if err != nil {
				fmt.Println("Error Occured ", err.Error())
				return nil
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

	sessionCmd.AddCommand(treeCmd)
	sessionCmd.AddCommand(interactCmd)
	sessionCmd.AddCommand(listCmd)
	sessionCmd.AddCommand(stopCmd)
	app.AddCommand(sessionCmd)
}
