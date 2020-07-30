package comms

import "io"

type Client interface {
	Send(Packet, int) (Packet, error)
	Shutdown()
}

type Conn interface {
	Close() error
	io.Writer
	io.Reader
}

type Dialer interface {
	Dial() Conn
}
