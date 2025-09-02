package autoroute

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

// ServerConn is the conn between the agent and gvisor (created in yamux.Client)
// ServerConn will recieve data from gvisor
// ServerConn data will be forwarded to beacon through http to yamux.Server
type ServerConn struct {
	ID       string
	Con net.Conn
}

func NewServerConn(id string, port string) (ServerConn, error) {
	sc := ServerConn{}
	conn, err := net.Dial("tcp", "127.0.0.1:"+port)
	if err != nil {
		fmt.Println("Error Connecting to local port", err)
		return sc, err
	}
	logrus.Warn("Accepted Connection")
	sc.Con = conn
	sc.ID = id

	return sc, nil
}

// TODO: Make this hang
func (c *ServerConn) Read(b []byte) (n int, err error) {
	//logrus.Info(c.ID, ": Someone is reading the server conn ", len(b), b[:30])

	return c.Con.Read(b)
}

func (c *ServerConn) Write(b []byte) (n int, err error) {
	//logrus.Info(c.ID, ": Someone is writing the server conn ", b)
	return c.Con.Write(b)
}

func (c *ServerConn) Close() error {
	//logrus.Info(c.ID, ": Someone is closing the server conn")
	return c.Con.Close()
}

func (c *ServerConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (c *ServerConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (c *ServerConn) SetDeadline(t time.Time) error      { return nil }
func (c *ServerConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *ServerConn) SetWriteDeadline(t time.Time) error { return nil }
