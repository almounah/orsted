package autoroute

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/nicocha30/ligolo-ng/cmd/proxy/config"
	"github.com/nicocha30/ligolo-ng/pkg/proxy"
	"github.com/nicocha30/ligolo-ng/pkg/proxy/netinfo"
	"github.com/nicocha30/ligolo-ng/pkg/proxy/netstack"
	"github.com/sirupsen/logrus"
)

type Route struct {
	RouteId       string
	Subnet        []string
	BeaconId      string
	PortNumber    string
	LocalListener net.Listener
	Proxy         *ServerConn
	ProxyMu       *sync.Mutex
	ProxyCond     *sync.Cond
	Agent         *ServerConn
	AgentMu       *sync.Mutex
	AgentCond     *sync.Cond
}

var ROUTE_LIST []*Route
var PORTCOUNT int = 0

func NewRoute(beaconId string, PortNumber string) (*Route, error) {
	r := Route{}
	r.RouteId = strconv.Itoa(len(ROUTE_LIST) + 1)
	r.BeaconId = beaconId
	r.PortNumber = PortNumber
	r.Subnet = []string{}

	var proxymu sync.Mutex
	r.ProxyMu = &proxymu
	r.ProxyCond = sync.NewCond(r.ProxyMu)

	var agentmu sync.Mutex
	r.AgentMu = &agentmu
	r.AgentCond = sync.NewCond(r.AgentMu)

	ROUTE_LIST = append(ROUTE_LIST, &r)
	PORTCOUNT++
	return &r, nil
}

func (r *Route) InitialiseTunInterface() error {

	tunName := "oss_" + r.BeaconId

	// Command 1: sudo ip tuntap add user parrot mode tun <tunName>
	cmd1 := exec.Command("ip", "tuntap", "add", "user", "parrot", "mode", "tun", tunName)
	cmd1Out, err := cmd1.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add tun device: %v\nOutput: %s", err, string(cmd1Out))
	}

	// Command 2: sudo ip link set <tunName> up
	cmd2 := exec.Command("ip", "link", "set", tunName, "up")
	cmd2Out, err := cmd2.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to set tun device up: %v\nOutput: %s", err, string(cmd2Out))
	}
	return nil
}

func (r *Route) AddSubnetForRoute(route string) error {

	tunName := "oss_" + r.BeaconId
	// Command 3: sudo ip route add <route> dev <tunName>
	cmd3 := exec.Command("ip", "route", "add", route, "dev", tunName)
	cmd3Out, err := cmd3.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add route: %v\nOutput: %s", err, string(cmd3Out))
	}

	r.Subnet = append(r.Subnet, route)

	return nil
}

func (r *Route) StartRoute() error {

	// Create Ligolo Stack
	tunName := "oss_" + r.BeaconId
	configState, err := config.GetInterfaceConfigState()
	if err != nil {
		return err
	}
	if ifaceConfig, ok := configState[tunName]; ok {
		if runtime.GOOS == "linux" && !ifaceConfig.Active {
			logrus.Debugf("Creating tun interface %s", tunName)
			if err := netinfo.CreateTUN(tunName); err != nil {
				logrus.Error(err)
			}
		}
	}

	ligoloStack, err := proxy.NewLigoloTunnel(netstack.StackSettings{
		TunName:     tunName,
		MaxInflight: 4096,
	})
	if err != nil {
		fmt.Println("Error Creating Tunnel", err)
		return err
	}

	if ifaceConfig, ok := configState[tunName]; ok {
		for _, ifcfg := range ifaceConfig.Routes {
			if !ifcfg.Active {
				logrus.Debugf("Creating route %s on interface %s", tunName, ifcfg.Destination)
				tun, err := netinfo.GetTunByName(tunName)
				if err != nil {
					logrus.Error(err)
					return err
				}
				if err := tun.AddRoute(ifcfg.Destination); err != nil {
					return err
				}
			}
		}
	}

	_, err = ligoloStack.GetStack().Interface().Name()
	if err != nil {
		logrus.Warn("unable to get interface name, err:", err)
	}

	// Start LocalListener
	l, err := net.Listen("tcp", "127.0.0.1:"+r.PortNumber)
	if err != nil {
		logrus.Info("Error Starting Listener ", err)
		return fmt.Errorf("failed to start listener: %w", err)
	}

	r.LocalListener = l
	logrus.Info("TCP listener started on 127.0.0.1:" + r.PortNumber)

	// Start go routine to set proxy client
	go func(a *Route) {
		for {
			conn, err := l.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					fmt.Println("Listener closed. Exiting accept loop.")
					break
				}
				fmt.Println("Error Accepting Connection ", err)
				continue // Optionally handle error appropriately
			}
			fmt.Println("Accepted connection successfully")
			proxyClient := &ServerConn{Con: conn, ID: "proxy_client"}
			a.Proxy = proxyClient
			a.ProxyCond.Signal()
		}
	}(r)

	// Start the Agent
	agentServer, err := NewServerConn("agent_server", r.PortNumber)
	if err != nil {
		logrus.Warn("Unable to create agent Server ", err)
	}
	r.Agent = &agentServer
	r.AgentCond.Signal()
	fmt.Println("Created Server_Conn, signalled Agent")
	fmt.Println(r.Agent)
	fmt.Println(agentServer)

	// Create yamux session
	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 120 * time.Second
	cfg.ConnectionWriteTimeout = 120 * time.Second
	r.ProxyMu.Lock()
	for r.Proxy == nil {
		r.ProxyCond.Wait()
	}
	r.ProxyMu.Unlock()

	session, err := yamux.Client(r.Proxy, nil)
	if err != nil {
		fmt.Println("Error in creating yamux client ", err)

	}
	_, err = session.Open()
	if err != nil {
		fmt.Println("Error in openeing yamux session ", err)

	}
	fmt.Println("Created yamux client")

	// Start Ligolo magic on server
	ctx, cancelTunnel := context.WithCancel(context.Background())
	// Handle packets
	go ligoloStack.HandleSession(session, ctx)

	// Check for closure
	go func() {
		// Check agent status
		for {
			select {
			case <-session.CloseChan(): // Agent closed
				fmt.Println("Closing session, need to delete route")
				logrus.WithFields(logrus.Fields{}).Warnf("Agent dropped.")
				cancelTunnel()
				time.Sleep(3*time.Second)
				err := r.DeleteRouteTunInterface()
				if err != nil {
					fmt.Println(err.Error())
				}
				r.DeleteRouteFromGlobalList()
				return
			}
		}
	}()

	return nil
}

func (r *Route) DeleteRouteFromGlobalList() error {
	for i, route := range ROUTE_LIST {
		if route == r {
			// Remove the element at index i
			ROUTE_LIST = append(ROUTE_LIST[:i], ROUTE_LIST[i+1:]...)
			break
		}
	}
	return nil
}

func (r *Route) DeleteRouteTunInterface() error {
	tunName := "oss_" + r.BeaconId
	// Command 3: sudo ip route add <route> dev <tunName>
	cmd3 := exec.Command("ip", "tuntap", "del", "dev", tunName, "mode", "tun")
	cmd3Out, err := cmd3.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to del tun: %v\nOutput: %s", err, string(cmd3Out))
	}

	return nil
}

func (r *Route) DeleteSubnetFromRoute(subnet string) error {
	var found bool = false
	for i, s := range r.Subnet {
		if s == subnet {
			// Remove target from slice
			r.Subnet = append(r.Subnet[:i], r.Subnet[i+1:]...)
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Subnet Not Found in route")
	}
	cmd3 := exec.Command("ip", "route", "del", subnet)
	cmd3Out, err := cmd3.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to del route: %v\nOutput: %s", err, string(cmd3Out))
	}

	return nil
}

func (r *Route) StopRoute() error {
	err := r.DeleteRouteTunInterface()
	r.LocalListener.Close()
	r.DeleteRouteFromGlobalList()
	return err
}
