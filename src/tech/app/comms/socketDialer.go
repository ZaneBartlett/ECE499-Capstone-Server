package comms

import (
	"net"
)

type UnixSocketDialer struct {
	socketName string
}

func NewUnixSocketDialer(socketName string) *UnixSocketDialer {
	return &UnixSocketDialer{socketName}
}

func (dialer *UnixSocketDialer) Dial() Conn {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: dialer.socketName, Net: "unix"})
	if err != nil {
		return nil
	} else {
		return conn
	}
}
