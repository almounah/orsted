package core

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"orsted/profiles"
	"orsted/protobuf/eventrpc"
	"orsted/protobuf/orstedrpc"
	"orsted/server/event"
	"orsted/server/utils"
	"orsted/server/wraprpc"
)


func StartServerGRPC(ip string, port int) {
    utils.PrintInfo("Starting Orsted Server on", fmt.Sprintf("%s:%d", ip, port))
    err := profiles.InitialiseProfile()
	if err != nil {
		log.Fatalf("failed to parse profile: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	maxRecvSize := 1024 * 1024 * 500 // 500 MB
	maxSendSize := 1024 * 1024 * 500 // 500 MB

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvSize),
		grpc.MaxSendMsgSize(maxSendSize),
	)


    utils.PrintInfo("Registering Orsted RPC Server")
	orstedrpc.RegisterOrstedRpcServer(grpcServer, &wraprpc.Server{})
    utils.PrintInfo("Registering Orsted Notification Server")
    eventrpc.RegisterNotifierServer(grpcServer, event.EventServerVar)
	grpcServer.Serve(lis)
}
