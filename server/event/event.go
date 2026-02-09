package event

import (
	"log"
	"orsted/protobuf/eventrpc"
	"sync"

)


type EventServer struct {
	eventrpc.UnimplementedNotifierServer
    mu      sync.Mutex
	Clients map[eventrpc.Notifier_SubscribeServer]struct{}
}

var EventServerVar = &EventServer{Clients: make(map[eventrpc.Notifier_SubscribeServer]struct{})}

func (s *EventServer) Subscribe(r *eventrpc.SubscribeRequest, stream eventrpc.Notifier_SubscribeServer) error {
    s.mu.Lock()
	s.Clients[stream] = struct{}{}
	s.mu.Unlock()

	// Keep the stream open
	<-stream.Context().Done()

	s.mu.Lock()
	delete(s.Clients, stream)
	s.mu.Unlock()

    return nil
}

func (s *EventServer) NotifyClients(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.Clients {
		if err := client.Send(&eventrpc.Notification{Message: message}); err != nil {
			log.Printf("Failed to send notification: %v", err)
			delete(s.Clients, client)
		}
	}
}
