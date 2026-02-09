package autoroute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/hashicorp/yamux"
	"github.com/nicocha30/ligolo-ng/pkg/protocol"
	"github.com/nicocha30/ligolo-ng/pkg/relay"
	"github.com/sirupsen/logrus"
)

type RPortForward struct {
	Id        int32
	RemoteSrc string
	LocalDst  string
}

type LigoloListener struct {
	ID      int32
	ctx     context.Context
	sess    *yamux.Session
	Conn    net.Conn
	addr    string
	network string
	to      string
}

func (l *LigoloListener) relayTCP() error {
	ligoloProtocol := protocol.NewEncoderDecoder(l.Conn)
	for {
		// Wait for BindResponses
		if err := ligoloProtocol.Decode(); err != nil {
			if err == io.EOF {
				// Listener closed.
				logrus.Debug("Listener closed connection (EOF)")
				return nil
			}
			return err
		}

		// We received a new BindResponse!
		response := ligoloProtocol.Payload.(*protocol.ListenerBindReponse)

		if err := response.Err; err != false {
			return errors.New(response.ErrString)
		}

		logrus.Debugf("New socket opened : %d", response.SockID)

		// relay connection
		go func(sockID int32) {
			forwarderSession, err := l.sess.Open()
			if err != nil {
				logrus.Error(err)
				return
			}

			forwarderProtocolEncDec := protocol.NewEncoderDecoder(forwarderSession)

			// Request socket access
			socketRequestPacket := protocol.ListenerSockRequestPacket{SockID: sockID}
			if err := forwarderProtocolEncDec.Encode(socketRequestPacket); err != nil {
				logrus.Error(err)
				return
			}
			// Get response back (ListenerSocketResponsePacket)
			if err := forwarderProtocolEncDec.Decode(); err != nil {
				logrus.Error(err)
				return
			}
			if err := forwarderProtocolEncDec.Payload.(*protocol.ListenerSockResponsePacket).Err; err != false {
				logrus.Error(forwarderProtocolEncDec.Payload.(*protocol.ListenerSockResponsePacket).ErrString)
				return
			}
			// If no error, establish TCP conn!
			logrus.Debugf("Listener relay established to %s (%s)!", l.to, l.network)

			// Dial the "to" target
			connFailed := false
			lconn, err := net.Dial(l.network, l.to)
			if err != nil {
				logrus.Error(err)
				connFailed = true
			}

			// Send connect ack (avoid races)
			connectionAckPacket := protocol.ListenerSocketConnectionReady{Err: connFailed}
			if err := forwarderProtocolEncDec.Encode(connectionAckPacket); err != nil {
				logrus.Error(err)
				return
			}

			if connFailed {
				return
			}

			// relay connections
			if err := relay.StartRelay(lconn, forwarderSession); err != nil {
				logrus.Error(err)
				return
			}
		}(response.SockID)

	}

}

func NewEmptyRouteForReverseForward(beaconId string, remoteSrc, localDst string) {
	r := Route{}
	r.RouteId = strconv.Itoa(len(ROUTE_LIST) + 1)
	r.BeaconId = beaconId
	r.Subnet = []string{}
	rportfwd := RPortForward{RemoteSrc: remoteSrc, LocalDst: localDst}
	r.ForwardedPort = []*RPortForward{&rportfwd}
	r.Active = false
	ROUTE_LIST = append(ROUTE_LIST, &r)
}

// Returning the ID from the beacon is needed to stop the listener later
func (r *Route) SendInstructionToRPortFwd(remoteSrc, localDst string) (int32, error) {
	conn, err := r.Session.Open()
	if err != nil {
		return -1, err
	}

	ligoloProtocol := protocol.NewEncoderDecoder(conn)

	// Request to open a new port on the agent
	listenerPacket := protocol.ListenerRequestPacket{Address: remoteSrc, Network: "tcp"}
	if err := ligoloProtocol.Encode(listenerPacket); err != nil {
		return -1, err
	}

	// Get response from agent
	if err := ligoloProtocol.Decode(); err != nil {
		return -1, err
	}
	response := ligoloProtocol.Payload.(*protocol.ListenerResponsePacket)
	if err := response.Err; err {
		return -1, errors.New(response.ErrString)
	}

	// Add the ID to the Object in the route
	for _, f := range r.ForwardedPort {
		if f.RemoteSrc == remoteSrc {
			f.Id = response.ListenerID
			break
		}
		
	}
	proxyListener := LigoloListener{ID: response.ListenerID, sess: r.Session, Conn: conn, addr: remoteSrc, network: "tcp", to: localDst}
		

	go func() {
		err := proxyListener.relayTCP()
		if err != nil {
			logrus.WithFields(logrus.Fields{"id": r.BeaconId}).Error("Listener relay failed with error: ", err)
			return
		}

		logrus.WithFields(logrus.Fields{"id": r.BeaconId}).Warning("Listener ended without error.")
		return
	}()
	return response.ListenerID, nil
}

func (r *Route) DeletePortFwdFromRoute(remoteSrc string) error {
	// Find the Id
	var toDelete *RPortForward
	for _, f := range r.ForwardedPort {
		if f.RemoteSrc == remoteSrc {
			toDelete = f
			break
		}
		
	}

	if toDelete == nil {
		return fmt.Errorf("Could not find rportfwd object ID. Check you specified correct remote src when attempting to delete")
	}

	// Open Yamux connection
	yamuxConnectionSession, err := r.Session.Open()
	if err != nil {
		return err
	}

	ligoloProtocol := protocol.NewEncoderDecoder(yamuxConnectionSession)

	// Send close request
	closeRequest := protocol.ListenerCloseRequestPacket{ListenerID: toDelete.Id}
	if err := ligoloProtocol.Encode(closeRequest); err != nil {
		return err
	}

	// Process close response
	if err := ligoloProtocol.Decode(); err != nil {
		return err

	}

	if err := ligoloProtocol.Payload.(*protocol.ListenerCloseResponsePacket).Err; err != false {
		return errors.New(ligoloProtocol.Payload.(*protocol.ListenerCloseResponsePacket).ErrString)
	}

	for i, rport := range r.ForwardedPort {
		if rport.Id == toDelete.Id {
			// Remove the element at index i
			r.ForwardedPort = append(r.ForwardedPort[:i], r.ForwardedPort[i+1:]...)
			break
		}
	}

	return nil
}
