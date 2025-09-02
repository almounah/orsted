package handler

import (
	"fmt"
	"os"
	"strings"

	"orsted/protobuf/orstedrpc"
	"orsted/server/event"
	"orsted/server/socks"
	"orsted/server/utils"
)

// Handle a task depending of the
// task type
func HandleTask(t *orstedrpc.Task, oldstate string, tasktype string) {
	switch tasktype {
	// It means we are dealing with a file
	// Don't display it, rathe download it to server
	case "respdownload":
		utils.PrintDebug("Received a Download Response")
		s := strings.Split(t.Command, " ")
		if len(s) < 2 {
			s := fmt.Sprintf("Big Bug in download")
			event.EventServerVar.NotifyClients(s)
			return
		}
		filepath := s[2]
		var f *os.File
		var err error
		if oldstate == "sent" {
			// Open new file
			f, err = os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		} else {
			f, err = os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		}
		if err != nil {
			utils.PrintDebug("Error in opening file", err.Error())
			event.EventServerVar.NotifyClients("Error in download " + err.Error())
			return
		}
		defer f.Close()
		if _, err := f.Write(t.Response); err != nil {
			utils.PrintDebug("Error in Writing response", err.Error())
			event.EventServerVar.NotifyClients("Error in download " + err.Error())
			return
		}
		res := fmt.Sprintf("\n >>> %s \n <<< %s \n", string(t.Command), string(t.State))
		event.EventServerVar.NotifyClients(res)

		// Means we are dealing with socks corner case
	case "respsocksunbind":
		// Signaling that socks got unbinded
		utils.PrintDebug("Received Unbind Response")
		// TODO: Find cleaner way instead of crashing go routine
		socks.SOCKS_LISTENER.Mu.Lock()
		socks.SOCKS_LISTENER.ToSendConn["unbind"] = nil
		socks.SOCKS_LISTENER.Cond.Signal() // Notify anyone waiting for a connection
		socks.SOCKS_LISTENER.Mu.Unlock()
		return
	case "respinteractiveshell":
		utils.PrintDebug("Received Interactive Shell response")
		if t.Response != nil && len(t.Response) != 0 {
			s := fmt.Sprintf(string(t.Response))
			event.EventServerVar.NotifyClients(s)
		}
        return

		// Just print to terminal
	default:
		s := fmt.Sprintf("\n >>> %s \n <<< %s \n", string(t.PrettyCommand), string(t.Response))
		event.EventServerVar.NotifyClients(s)
	}
}
