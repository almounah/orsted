package autoroute

import (
	"context"
	"net/http"
	"orsted/server/utils"
	"strings"

	"github.com/coder/websocket"
)

func HandleAutorouteWebsocket(w http.ResponseWriter, r *http.Request) {
	utils.PrintInfo("Received Websocket Autoroute")
    ws, err := websocket.Accept(w, r, nil)
    if err != nil {
		utils.PrintDebug("Error Accepting websocket --> ", err)
        return
    }
    netctx := context.Background()

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	id := parts[len(parts)-1]
    utils.PrintDebug("ID =", id)

	ProxyConn := websocket.NetConn(netctx, ws, websocket.MessageBinary)
	ActivateRoute(id, ProxyConn)

	return
}

