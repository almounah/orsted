package autoroute

import (
	"fmt"
	"orsted/protobuf/orstedrpc"
	"strings"
)

func AddRoute(beacondId string, subnet string) error {
	// Check if route already exists
	var r *Route
	for _, c := range ROUTE_LIST {
		if c.BeaconId == beacondId {
			r = c
		}
	}

	// If it is just add the subnet
	if r != nil {
		err := r.AddRouteToTun(subnet)
		if err != nil {
			fmt.Println("Error ", err)
			return err
		}
		r.Subnet = append(r.Subnet, subnet)
		return fmt.Errorf("Beacon already ligoloing, will only add route locally")
	}

	// Otherwise create Empty Route that will be populated if Websocket Success
	NewEmptyRoute(beacondId, subnet)
	

	return nil
}


func ListRoute() *orstedrpc.RouteList {
	var ListOfRoute []*orstedrpc.Route

	for _, c := range ROUTE_LIST {
		var route orstedrpc.Route
		route.RouteId = c.RouteId
		route.BeaconId = c.BeaconId
		route.Subnet = strings.Join(c.Subnet, ", ")
		// Pretty Print Forwarded Port
		var RportfwdString []string 
		for i := 0; i < len(c.ForwardedPort); i++ {
			RportfwdString = append(RportfwdString, c.ForwardedPort[i].LocalDst + "<-->" + c.ForwardedPort[i].RemoteSrc)
			
		}
		route.Rportfwd = strings.Join(RportfwdString, ", ")
		ListOfRoute = append(ListOfRoute, &route)
	}

	var routeList orstedrpc.RouteList
	routeList.Routes = ListOfRoute
	return &routeList
}

func DeleteSubnetRoute(beaconId, subnet string) error {
	var r *Route
	for _, c := range ROUTE_LIST {
		if c.BeaconId == beaconId {
			r = c
		}
	}
	if r == nil {
		return fmt.Errorf("Beacon Not found in list of route, maybe it died")
	}
	err := r.DeleteSubnetFromRoute(subnet)
	if err != nil {
		return err
	}
	if len(r.Subnet) == 0 {
		r.StopRoute()
	}
	return nil
	
}

func AddReversePortForwardInRoute(beacondId string, remoteSrc, localDst string) error {
	// Check if route already exists
	var r *Route
	for _, c := range ROUTE_LIST {
		if c.BeaconId == beacondId {
			r = c
		}
	}

	// If it is just send the command to Rev PortForward
	if r != nil {
		err := r.SendInstructionToRPortFwd(remoteSrc, localDst)
		if err != nil {
			fmt.Println("Error ", err)
			return err
		}
		r.ForwardedPort = append(r.ForwardedPort, RPortForward{RemoteSrc: remoteSrc, LocalDst: localDst})
		return fmt.Errorf("Beacon already ligoloing, will instruct to rportfwd")
	}

	// Otherwise create Empty Route that will be populated if Websocket Success
	NewEmptyRouteForReverseForward(beacondId, remoteSrc, localDst)
	

	return nil
}



