package main

import (
	"flag"
	"net/http"
	"runtime"
	"tech/app/comms"
	"tech/app/logger"

	"github.com/go-chi/chi"
)

const (
	socketName = "@/tmp/socketTest.sock"
)

// Env is a container for objects that may be overwritten by tests
type Env struct {
	client comms.Client
}

var env *Env

func main() {

	var httpLog bool
	var logNormal bool
	var logDebug bool

	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.BoolVar(&httpLog, "h", false, "Log http requests")
	flag.BoolVar(&logNormal, "l", false, "Logs additional application statements")
	flag.BoolVar(&logDebug, "d", false, "Logs debug statements")
	flag.Parse()

	env = &Env{}

	logger.Init("server")
	logger.LogToStdout = false
	if logNormal {
		logger.LogToStdout = true
	}
	if logDebug {
		logger.Debug = true
	}

	env.client = createSocketClient()
	defer env.client.Shutdown()

	router := chi.NewRouter()
	configureRoutes(router, httpLog)

	logger.Log("Starting http server")
	http.ListenAndServe(":8080", router)
}

func createSocketClient() *comms.SocketClient {
	dialer := comms.NewUnixSocketDialer(socketName)
	client := comms.NewClient(dialer)
	return client
}
