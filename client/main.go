package main

import (
	"flag"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"orsted/client/grumblecli"
	"orsted/client/notificationhandle"
)

func connectGrpc(ip string, port string) grpc.ClientConnInterface {
	conn, err := grpc.NewClient(ip+":"+port, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
		    grpc.MaxCallRecvMsgSize(1024*1024*500),
		    grpc.MaxCallSendMsgSize(1024*1024*500),
	    ))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	fmt.Println("Connection Worked")
	return conn
}

func PrintBanner() {
	red := "\033[31m"  // ANSI escape code for red
	reset := "\033[0m" // Reset color

	asciiArt := `
                                 @@@@@@@                       
                             @@@@@@@@@@@@@                  
                            @@@@@@@@@@@@@@@                 
            @@             @@@@@@@@@@@@@@@@@@@@@@           
            @@            @@@@@@@@@@@@@@@@@@@@@@@@          
            @@          @@@@@@@@@@@@@@@@@@@@@@@@@@@         
           @@@      @@@@@@@@@@@@@@       @@@@@@@@@@@@@@@ @@@
          @@@      @@@@@@@@@@@@@  @@@@@       @@@@@@@@@@@@  
        @@@@      @@@@@@@@@@@@   @@@@@@@@@@                 
       @@@@  @   @@@@@@@@@     @@@@@@@@@@@@@@@              
        @@@ @@@@@@@@@@@@@     @@@@@@@@@@@@@@@@@@@           
        @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@          
        @@@@  @@    @@@@@@@@@@@@@@@@@@@@@   @@@@@@@@        
       @@@@@@@@@@@@          @@@@@@@@@@         @@@@@       
        @@@@@@@@@@@@@@@@@@                        @@@@      
          @@@@@@@@@@@@@@@@@    @                    @@      
             @@@    @@@@@@@  @@@                     @@     
               @    @@@@@@   @@@                            
                  @@@@@@      @@@@@                         
                  @@@@@@       @@@@@@@@                     
                  @@@@@@@@@   @@@                           
                      @@@@@@@@@@                            
                             @@                             
    `

	fmt.Println(red + asciiArt + reset)
}

func main() {
	PrintBanner()
	ip := flag.String("ip", "0.0.0.0", "server grpc ip")
	port := flag.String("port", "50051", "server grpc port")
	confPath := flag.String("conf", "./data/clientconf.toml", "conf file default path")
	flag.Parse()

	err := grumblecli.ParseClientConf(*confPath)
	if err != nil {
		fmt.Println("Error while parsing client conf", err.Error())
		return
	}

	conn := connectGrpc(*ip, *port)
	go notificationhandle.HandleNotfication(conn)
	grumblecli.SetCommands(conn)
	grumblecli.Run()
}
