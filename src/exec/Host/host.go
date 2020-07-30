package main

import (
	"flag"
	"net"
	"tech/app/comms"
	"tech/app/logger"
	"tech/mixer"
)

const (
	logfileName = "Host.log"
	socketName  = "@/tmp/socketTest.sock"
)

var gitHash string
var compileDate string

func main() {

	var logNormal bool
	var logDebug bool

	flag.BoolVar(&logNormal, "l", false, "Logs additional application statements")
	flag.BoolVar(&logDebug, "d", false, "Logs debug statements")
	flag.Parse()

	logger.Init("Host")
	logger.LogToStdout = false
	if logNormal {
		logger.LogToStdout = true
	}
	if logDebug {
		logger.Debug = true
	}

	mixerDev := mixer.NewMixer()
	err := mixerDev.Start()
	if err != nil {
		logger.Log("Failed to initialize subsystems, error is %v, exiting", err)
		return
	}

	host, err := createSocketHost()
	if err != nil {
		logger.Log("Failed to create socket host, error is %v, exiting", err)
		return
	}

	handleClientRequest(host.Out, host.In, mixerDev)
}

func createSocketHost() (*comms.SocketHost, error) {
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socketName, Net: "unix"})
	if err != nil {
		logger.Log("Failed to generate listener, err is %v", err)
		return nil, err
	}
	host := comms.NewHost()
	go host.Listen(listener)
	return host, nil
}

func handleClientRequest(in chan comms.Packet, out chan comms.Packet, dev *mixer.Mixer) {
	for {
		packet := <-in
		switch packet.Header.Target {
		default:
			response, err := dev.Action(packet.Header.Target, packet.Header.Action, packet.Data)
			if err != nil {
				logger.Log("Unrecognized target received, '%s', error is '%v'", packet.Header.Target, err)
			}

			out <- comms.BuildResponsePacket(packet.Header, response)
		}
	}
}
