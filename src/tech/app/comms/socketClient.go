package comms

import (
	"container/list"
	"encoding/json"
	"fmt"
	"sync"
	"tech/app/logger"
	"time"
)

type request struct {
	msgID        uint32
	responseChan chan Packet
}

type SocketClient struct {
	dialer       Dialer
	Connected    bool
	conn         Conn
	enc          *json.Encoder
	messageQueue *list.List
	wg           sync.WaitGroup
	msgCounter   uint32
}

func NewClient(dialer Dialer) *SocketClient {
	var client SocketClient
	client.messageQueue = list.New()
	client.msgCounter = 1
	client.dialer = dialer
	go client.doDial()
	return &client
}

// Shutdown the connection
func (client *SocketClient) Shutdown() {
	if client.conn != nil {
		client.conn.Close()
		logger.Log("Closed connection")
	}
}

func nextDelay(delay int64) int64 {
	if delay == 0 {
		delay = 200
	} else {
		if delay < 1000 {
			delay += 200
		} else if delay < 4000 {
			delay += 500
		} else if delay < 10000 {
			delay += 2000
		} else if delay < 60000 {
			delay += 5000
		}
	}
	return delay
}

func (client *SocketClient) doDial() {
	if client.dialer == nil {
		logger.Log("Unable to dial, dialer is nil")
		return
	}
	connectDelay := int64(0)
	for {
		//ok := false
		conn := client.dialer.Dial()
		if conn != nil {
			// Do socket client
			client.conn = conn
			client.enc = json.NewEncoder(conn)
			client.Connected = true
			client.doClientReceive()
			client.Connected = false
			connectDelay = 0
		} else {
			//if !ok {
			connectDelay = nextDelay(connectDelay)
			logger.Log("connectDelay delay: %v mSec\r\n", connectDelay)
			time.Sleep(time.Duration(connectDelay) * time.Millisecond)
		}
	}
}

func (client *SocketClient) incrementMsgCounter() {
	client.msgCounter++
	// 0 is an invalid value
	if client.msgCounter == 0 {
		client.msgCounter++
	}
}

// Send - Send a packet to the host and wait for response
func (client *SocketClient) Send(packet Packet, timeout int) (Packet, error) {
	var result Packet
	// if packet.Header.MsgId == 0 {
	// 	return result, fmt.Errorf("Invalid send packet, msg id can't be 0")
	// }
	if !client.Connected {
		return result, fmt.Errorf("Client is not connected to host, unable to send")
	}
	packet.Header.MsgId = client.msgCounter
	client.incrementMsgCounter()
	responseChan := make(chan Packet)

	client.messageQueue.PushBack(request{msgID: packet.Header.MsgId, responseChan: responseChan})
	err := client.enc.Encode(packet)
	if err != nil {
		logger.Log("Encode err %v - Exiting", err.Error())
		client.findAndRemoveMessage(packet.Header.MsgId)
		return result, err
	}
	//client.logger.Printf("Added message id %d", packet.Header.MsgId)
	if timeout == 0 {
		result = <-responseChan
	} else {
		select {
		case result = <-responseChan:
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			return result, fmt.Errorf("Timed out waiting for response")
		}
	}

	if !result.Header.Ack {
		logger.Log("Server replied but did not ack packet")
		return result, fmt.Errorf("Server replied but did not ack packet")
	}

	return result, nil
}

func (client *SocketClient) findAndRemoveMessage(msgID uint32) (request, error) {
	var result request
	for e := client.messageQueue.Front(); e != nil; e = e.Next() {
		//var nilObj request
		if e.Value != nil {
			req := e.Value.(request)
			if req.msgID == msgID {
				//client.logger.Printf("Found message id %d", msgID)
				if result.msgID == 0 {
					client.messageQueue.Remove(e)
					result = req
				} else {
					logger.Log("Found duplicate message in queue, deleting duplicate")
				}
			}
		}
	}
	if result.msgID != 0 {
		return result, nil
	}
	return result, fmt.Errorf("Unable to locate request message id %d", msgID)
}

func (client *SocketClient) doClientReceive() {
	client.wg.Add(1)
	defer client.wg.Done()
	dec := json.NewDecoder(client.conn)
	for {
		var packet Packet
		err := dec.Decode(&packet)
		if err == nil && packet.Header.MsgId != 0 {
			request, err2 := client.findAndRemoveMessage(packet.Header.MsgId)
			if err2 != nil {
				logger.Log("Error finding request, error is %v", err2)
			} else {
				request.responseChan <- packet
			}
		} else if err != nil {
			logger.Log("Read err %v - exiting", err.Error())
			break
		} else if packet.Header.MsgId == 0 {
			logger.Log("Received packet with invalid message id, ignoring")
		} else {
			logger.Log("Read err %v - exiting", err.Error())
			break
		}
	}
	logger.Log("Receive exiting")
}
