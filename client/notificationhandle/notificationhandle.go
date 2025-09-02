package notificationhandle

import (
	"context"
	"io"
	"log"

	"google.golang.org/grpc"

	"orsted/protobuf/eventrpc"

	"github.com/fatih/color"
)

func HandleNotfication(conn grpc.ClientConnInterface) {
	eventNotifierClient := eventrpc.NewNotifierClient(conn)

	ctx := context.Background()

	stream, err := eventNotifierClient.Subscribe(ctx, &eventrpc.SubscribeRequest{})
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

    for {
		notification, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive notification: %v", err)
		}
        bold := color.New(color.FgCyan).Add(color.Bold)
	    bold.Print(notification.Message)

	}
}
