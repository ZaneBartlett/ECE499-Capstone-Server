package logger

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

const (
	localLogFileDir = "/var/log/"
)

var localLogger *log.Logger
var syslogLogger *log.Logger
var debugLogger *log.Logger
var stdoutLogger *log.Logger

// LogtoLocal enables writing logs to a local file
var LogtoLocal bool

// LocalPath returns the location of the local file
var LocalPath string

// LogToStdout enables printing all logs to the std out
var LogToStdout bool

// LogtoSyslog enables writing all logs to the syslog
var LogtoSyslog bool

// Debug enables the debug logger
var Debug bool

// Init prepares the log files
func Init(appName string) {
	filename := localLogFileDir + appName + ".log"
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to open logfile %s\r\n", filename)
		return
	}
	localLogger = log.New(file, "", log.Lshortfile|log.LUTC|log.Ldate|log.Ltime)

	createSyslogLogger(appName)

	stdoutLogger = log.New(os.Stdout, "", log.Lshortfile|log.LUTC|log.Ldate|log.Ltime)

	debugFilename := localLogFileDir + appName + "_debug" + ".log"
	debugFile, err := os.Create(debugFilename)
	if err != nil {
		fmt.Printf("Failed to open debug logfile %s\r\n", debugFilename)
		return
	}
	debugLogger = log.New(debugFile, "", log.Lshortfile|log.LUTC|log.Ldate|log.Ltime)

	LogtoLocal = false
	LogToStdout = false
	LogtoSyslog = true
	Debug = false
}

func createSyslogLogger(appName string) {
	var err error
	syslogLogger, err = syslog.NewLogger(syslog.LOG_NOTICE, log.Lshortfile|log.LUTC|log.Ldate|log.Ltime)
	if err != nil {
		fmt.Printf("Failed to create syslog writer, err is %v\r\n", err)
		syslogLogger = nil
		return
	}
}

// Log writes to the default log file
func Log(format string, v ...interface{}) {
	if LogtoLocal && localLogger != nil {
		localLogger.Output(2, fmt.Sprintf(format, v...))
	}

	if LogtoSyslog && syslogLogger != nil {
		syslogLogger.Output(2, fmt.Sprintf(format, v...))
	}

	if LogToStdout && stdoutLogger != nil {
		stdoutLogger.Output(2, fmt.Sprintf(format, v...))
	}
}

// LogDebug writes to the default debug log
func LogDebug(format string, v ...interface{}) {
	if !Debug {
		return
	}
	if debugLogger != nil {
		debugLogger.Output(2, fmt.Sprintf(format, v...))
	}

	if LogToStdout && stdoutLogger != nil {
		stdoutLogger.Output(2, fmt.Sprintf(format, v...))
	}
}
