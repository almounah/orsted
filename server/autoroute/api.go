package autoroute

import (
	"fmt"
	"orsted/protobuf/orstedrpc"
	"strconv"
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
		err := r.AddSubnetForRoute(subnet)
		if err != nil {
			fmt.Println("Error ", err)
			return err
		}
		return fmt.Errorf("Beacon already ligoloing, will only add route locally")
	}

	// Otherwise create a new route
	port := 33433 + PORTCOUNT

	r, err := NewRoute(beacondId, strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error ", err)
		return err
	}
	err = r.InitialiseTunInterface()
	if err != nil {
		fmt.Println("Error ", err)
		// Failed to initialise, delete from global
		r.DeleteRouteFromGlobalList()
		return err
	}

	err = r.AddSubnetForRoute(subnet)
	if err != nil {
		fmt.Println("Error ", err)
		// Failed to add 1 route, delete from global
		r.DeleteRouteTunInterface()
		r.DeleteRouteFromGlobalList()
		return err
	}

	go r.StartRoute()
	return nil
}


func ListRoute() *orstedrpc.RouteList {
	var ListOfRoute []*orstedrpc.Route

	for _, c := range ROUTE_LIST {
		var route orstedrpc.Route
		route.RouteId = c.RouteId
		route.BeaconId = c.BeaconId
		route.Subnet = strings.Join(c.Subnet, ", ")
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
	r.DeleteSubnetFromRoute(subnet)
	if len(r.Subnet) == 0 {
		r.StopRoute()
	}
	return nil
	
}
