package wraprpc

import (
	"net"
)

func isTcpAddressInUse(address string) bool {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		// If there's an error, the address is likely in use or invalid.
		return true
	}
	// Close the listener if the address is available.
	ln.Close()
	return false
}
