package system

import (
	"net"
	"strconv"
	"strings"
)

// FindOpenPort returns the number of an available port.
//
// Returns an error if something goes wrong.
func FindOpenPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	address := listener.Addr().String()
	portStr := address[strings.LastIndex(address, ":")+1:]
	port, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(port), nil
}
