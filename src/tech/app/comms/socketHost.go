package comms

import (
	"encoding/json"
	"io"
	"net"
	"tech/app/logger"
)

// SocketHost listens for connections and for each connection spawns a receive and response func that correspond
//  with Out and In channels respectively
type SocketHost struct {
	Connected bool
	Ready     bool
	Out       chan Packet
	In        chan Packet
}

// NewHost returns a new SocketHost
func NewHost() *SocketHost {
	host := SocketHost{}
	host.Out = make(chan Packet)
	host.In = make(chan Packet)
	return &host
}

// Listen monitors the socket and handles all connections
func (host *SocketHost) Listen(listener net.Listener) {
	exitFlag := false
	for !exitFlag {
		host.Ready = true
		logger.Log("Host Listener Ready")
		socketConn, err := listener.Accept()
		host.Ready = false
		if err != nil {
			logger.Log("Could not accept connection, error is %v, exiting", err.Error())
			exitFlag = true
			continue
		} else if socketConn == nil {
			continue
		}
		logger.Log("Host Listener connection accepted")

		exitDetect := make(chan bool)
		logger.Log("Host Listener new connection")
		host.Connected = true
		go host.doHostResponse(socketConn, exitDetect)
		host.doHostReceive(socketConn, exitDetect)
	}
}

func (host *SocketHost) doHostReceive(conn Conn, exit chan bool) {
	logger.Log("Host receive starting")
	dec := json.NewDecoder(conn)
	for {
		var packet Packet
		err := dec.Decode(&packet)
		if err == nil && packet.Header.MsgId != 0 {
			host.Out <- packet
		} else if err == io.EOF {
			logger.Log("Host receive connection closed, exiting")
			exit <- true
			break
		} else if packet.Header.MsgId == 0 {
			logger.Log("Received packet with no header, ignoring")
		} else {
			logger.Log("Failed to decode byte stream, err is %v, ignoring", err.Error())
		}
	}
	logger.Log("Host receive exiting")
	host.Connected = false
}

func (host *SocketHost) doHostResponse(conn Conn, exit chan bool) {
	logger.Log("Host response starting")
	exitFlag := false
	enc := json.NewEncoder(conn)
	for !exitFlag {
		select {
		case val := <-host.In:
			err := enc.Encode(val)
			if err != nil {
				logger.Log("Failed to encode packet id %d, error is %v, ignoring", val.Header.MsgId, err.Error())
			}
		case <-exit:
			logger.Log("Host response exit request received")
			exitFlag = true
			break
		}
	}
	logger.Log("Host response exiting")
}
