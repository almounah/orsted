package autoroute

import (
	"context"
	"fmt"
	"net"
	"orsted/server/utils"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
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
	ForwardedPort []RPortForward
	BeaconId      string
	ProxyConn     net.Conn
	Session       *yamux.Session
	Active        bool
}

var ROUTE_LIST []*Route
var PORTCOUNT int = 0

func NewEmptyRoute(beaconId string, subnet string) {
	r := Route{}
	r.RouteId = strconv.Itoa(len(ROUTE_LIST) + 1)
	r.BeaconId = beaconId
	r.Subnet = []string{subnet}
	r.Active = false
	ROUTE_LIST = append(ROUTE_LIST, &r)
}

func ActivateRoute(beaconId string, wsConn net.Conn) error {
	var r *Route
	for _, c := range ROUTE_LIST {
		if c.BeaconId == beaconId {
			r = c
		}
	}
	if r == nil {
		return fmt.Errorf("Error in design, Activating a non existant Empty Route")
	}
	r.ProxyConn = wsConn

	if len(r.Subnet) != 1 && len(r.Subnet) != 0 {
		return fmt.Errorf("Error in design, Activating a route with multiple subnet ?!")
	}

	err := r.InitialiseTunInterface()
	if err != nil {
		fmt.Println("Error ", err)
		// Failed to initialise, delete from global
		r.DeleteRouteFromGlobalList()
		return err
	}

	// Start a subnet if the route was create for that. In that case r.Subnet will contain the only subnet
	if len(r.Subnet) == 1 {
		err = r.AddRouteToTun(r.Subnet[0])
		if err != nil {
			r.StopRoute()
			return err
		}
	}

	r.StartRoute()
	r.Active = true

	// Start revportfwd if the route was create for that. In that case r.ForwardedPort will contain the only Port to rev forward
	if len(r.ForwardedPort) == 1 {
		err = r.SendInstructionToRPortFwd(r.ForwardedPort[0].RemoteSrc, r.ForwardedPort[0].LocalDst)
		if err != nil {
			r.StopRoute()
			return err
		}
	}
	return nil
}

func (r *Route) InitialiseTunInterface() error {

	tunName := "oss_" + r.BeaconId

	// Command 1: sudo ip tuntap add user parrot mode tun <tunName>

	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to set tun device up %v", err)
	}
	cmd1 := exec.Command("ip", "tuntap", "add", "user", u.Username, "mode", "tun", tunName)
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

func (r *Route) AddRouteToTun(route string) error {
	tunName := "oss_" + r.BeaconId
	// Command 3: sudo ip route add <route> dev <tunName>
	cmd3 := exec.Command("ip", "route", "add", route, "dev", tunName)
	cmd3Out, err := cmd3.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add route: %v\nOutput: %s", err, string(cmd3Out))
	}

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

	// Create yamux session
	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 120 * time.Second
	cfg.ConnectionWriteTimeout = 120 * time.Second

	session, err := yamux.Client(r.ProxyConn, nil)
	r.Session = session
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
				time.Sleep(3 * time.Second)
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
	utils.PrintDebug("Deleting Route From GLobal List -> ", r.RouteId, ROUTE_LIST)
	for i, route := range ROUTE_LIST {
		if route.RouteId == r.RouteId {
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
		utils.PrintDebug("Iterating Over Route --> ", i, s)
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
	r.ProxyConn.Close()
	r.DeleteRouteFromGlobalList()
	return err
}
