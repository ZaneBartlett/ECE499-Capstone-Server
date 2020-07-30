package localsocket

import (
	"io"
	"net"
	"sync"
	"tech/app/logger"
	"time"
)

// Host is a struct for creating and listening on a unix socket
type Host struct {
	Ready      bool
	socketName string
	Out        chan []byte
	In         chan []byte
	conn       *net.UnixConn
	exitChan   chan bool
	wg         sync.WaitGroup
}

// NewHost returns a socket listener that creates and listens on a unix socket
func NewHost(socketName string) *Host {
	var host Host
	host.Ready = false
	host.socketName = socketName
	host.Out = make(chan []byte)
	host.In = make(chan []byte)
	host.conn = nil
	host.exitChan = make(chan bool)
	return &host
}

// Shutdown the connection
func (host *Host) Shutdown() {
	host.exitChan <- true
}

// Listen opens a unix socket and sends and receives data via the rx cand tx channels
func (host *Host) Listen() error {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: host.socketName, Net: "unix"})
	defer listener.Close()
	if err != nil {
		logger.Log("Could not listen on socket %s ERROR: %s", host.socketName, err.Error())
		return err
	}
	connChan := make(chan *net.UnixConn)
	logger.Log("Listening on socket %s", host.socketName)
	go host.doConnectionListener(listener, connChan)
	exitFlag := false
	for !exitFlag {
		var socketConn *net.UnixConn
		select {
		case socketConn = <-connChan:
			exitDetect := make(chan bool)
			logger.Log("New Connection on %v", socketConn.LocalAddr())
			go host.doHostReceive(socketConn, exitDetect)
			go host.doHostResponse(socketConn, exitDetect)
		case <-host.exitChan:
			logger.Log("Exit requested")
			exitFlag = true
			break
		}
	}
	return nil
}

func (host *Host) doConnectionListener(listener *net.UnixListener, connChan chan *net.UnixConn) {
	for {
		host.Ready = true
		logger.Log("LISTENER READY")
		unixConn, err := listener.AcceptUnix()
		if err != nil {
			logger.Log("Could not accept connection. ERROR: %s\n\n", err.Error())
			return
		} else if unixConn == nil {
			continue
		}
		connChan <- unixConn
	}
}

func (host *Host) doHostReceive(unixConn *net.UnixConn, exit chan bool) {
	buf := make([]byte, 1024)
	for {
		n, err := (*unixConn).Read(buf)
		if err == nil && n > 0 {
			newBuf := buf[0:n]
			host.Out <- newBuf
			//fmt.Printf("Ip: Received %d bytes\r\n", len(newBuf))
		} else if err == io.EOF {
			logger.Log("Listener - Connection closed, exiting")
			exit <- true
			break
		} else {
			logger.Log("Listener error exiting, err is %v", err.Error())
			exit <- true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	logger.Log("Listener - Exiting receive")
}

func (host *Host) doHostResponse(unixConn *net.UnixConn, exit chan bool) {
	exitFlag := false
	for !exitFlag {
		select {
		case val := <-host.In:
			//fmt.Printf("Ip client - Sending %d bytes\r\n", len(val))
			_, err := unixConn.Write(val)
			if err != nil {
				logger.Log("Listener Error: %v. Exiting", err.Error())
				exitFlag = true
				break
			}
		case <-exit:
			logger.Log("Listener - Exit request received, exiting transmit")
			exitFlag = true
			break
		}
	}
	logger.Log("Listener - Exiting response")
}
