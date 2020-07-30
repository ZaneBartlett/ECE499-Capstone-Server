package localsocket

import (
	"net"
	"sync"
	"tech/app/logger"
	"time"
)

// Client is a struct for connecting to a unix socket
type Client struct {
	Connected    bool
	socketName   string
	Out          chan []byte
	In           chan []byte
	conn         *net.UnixConn
	sendExitChan chan bool
	wg           sync.WaitGroup
}

// NewClient returns a struct for connecting and transferring data on a unix socket
func NewClient(socketName string) *Client {
	var client Client
	client.Connected = false
	client.socketName = socketName
	client.Out = make(chan []byte)
	client.In = make(chan []byte)
	client.sendExitChan = make(chan bool)
	client.conn = nil
	return &client
}

// Shutdown the connection
func (client *Client) Shutdown() {
	if client.conn != nil {
		client.conn.CloseRead()
		client.conn.Close()
		logger.Log("Closed connection")
	}
}

// Connect to the specified unix socket and send and receive via provided channels
func (client *Client) Connect() error {
	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: client.socketName, Net: "unix"})
	if err != nil {
		logger.Log("Could not connect to: %v ERROR: %s", client.socketName, err.Error())
		client.conn = nil
		return err
	}
	client.conn = conn
	logger.Log("Connected to: %v", client.socketName)
	go client.doClientReceive()
	go client.doClientSend()
	return nil
}

func (client *Client) doClientReceive() {
	client.wg.Add(1)
	defer client.wg.Done()
	buf := make([]byte, 1024)
	for {
		n, err := (*client.conn).Read(buf)
		if err == nil && n > 0 {
			//fmt.Printf("Ip: Received %d bytes\r\n", n)
			newBuf := make([]byte, n)
			copy(newBuf, buf[0:n])
			client.Out <- newBuf
		} else {
			logger.Log("Read err %v - exiting", err.Error())
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	logger.Log("Receive exiting")
}

func (client *Client) doClientSend() {
	client.wg.Add(1)
	defer client.wg.Done()
	exitFlag := false
	for !exitFlag {
		select {
		case val := <-client.In:
			//fmt.Printf("Sending %d bytes\r\n", len(val))
			numBytes, err := client.conn.Write(val)
			if err != nil {
				logger.Log("Write err %v - Exiting", err.Error())
				exitFlag = true
				break
			}
			if numBytes != len(val) {
				logger.Log("Failed to send sufficient bytes, requested %d, sent %d - Exiting", len(val), numBytes)
				exitFlag = true
				break
			}
		case <-client.sendExitChan:
			logger.Log("Exit request received")
			exitFlag = true
			break
		}

	}
	logger.Log("Transmit exiting")
}
