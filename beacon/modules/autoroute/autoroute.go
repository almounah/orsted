package autoroute

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/yamux"

	"orsted/beacon/modules/autoroute/agent"
	"orsted/beacon/utils"
)



func StartAutorouting(ctx context.Context) error {
	fmt.Println("Sleeping for 5 seconds")
	fmt.Println("This is updated beacon")

	netConn, err := utils.ParentPeer.GetRealTimeConn(utils.CurrentBeaconId)
	if err != nil {
		fmt.Println("Error Getting Real Time Conn ", err)
		return err
	}

	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 120 * time.Second // ping every 2min
	cfg.ConnectionWriteTimeout = 120 * time.Second

	yamuxConn, err := yamux.Server(netConn, cfg)
	if err != nil {
		fmt.Println("Error Creating Server ", err)
		return err
	}
	fmt.Println("Started Yamux Server Successfully")

	for {
		conn, err := yamuxConn.Accept()
		if err != nil {
			fmt.Println("Error Accepting Yamux Conn", err)
			return err
		}
		utils.Print("Accepted Yamux Session successfully, will start handling")
		go agent.HandleConn(conn)
	}

}

func StopSocksServer() {
}
