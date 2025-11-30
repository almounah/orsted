package customtcp

//Useless

import (
	"net"
	"orsted/beacon/utils"
	"time"
)

// Used for RealtimeConn for TCP PIVOT
// It just add a marker at the beginning 2
// That way it is distinguished in pivot realtime
type CustomNetConn struct {
	neConn net.Conn
	marker byte
}

func NewCustomNetConn(c net.Conn, marker byte) *CustomNetConn {
    return &CustomNetConn{
        neConn: c,
        marker: marker,
    }
}

// Write prefixes the data with the marker byte.
func (c *CustomNetConn) Write(b []byte) (int, error) {
    // Allocate a single buffer: [marker][data...]
	utils.Print("In Write for TCP Custom COnn")

	towrite := append([]byte{c.marker}, b...)
	utils.Print("Will Write ---> \n", towrite)
	utils.Print("Intead of writing ---> \n", b)
    n, err := c.neConn.Write(towrite)
	utils.Print("Done Writing", n)

    // Adjust the number of bytes reported as written (exclude marker)
    if n > 0 {
        return n - 1, err
    }

    return 0, err
}

// Read simply proxies reads.
func (c *CustomNetConn) Read(b []byte) (int, error) {
	utils.Print("In Read for TCP Custom COnn")
	res, err := c.neConn.Read(b)
	utils.Print("Done Reading --> ", b)
    return res, err
}

// Remaining Conn methods:
func (c *CustomNetConn) Close() error {
	utils.Print("In Close for TCP Custom COnn")
    return c.neConn.Close()
}

func (c *CustomNetConn) LocalAddr() net.Addr {
    return c.neConn.LocalAddr()
}

func (c *CustomNetConn) RemoteAddr() net.Addr {
    return c.neConn.RemoteAddr()
}

func (c *CustomNetConn) SetDeadline(t time.Time) error {
    return c.neConn.SetDeadline(t)
}

func (c *CustomNetConn) SetReadDeadline(t time.Time) error {
    return c.neConn.SetReadDeadline(t)
}

func (c *CustomNetConn) SetWriteDeadline(t time.Time) error {
    return c.neConn.SetWriteDeadline(t)
}
