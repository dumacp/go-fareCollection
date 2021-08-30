package logstrans

import (
	"io/ioutil"
	"log"
	"os"
)

//LogError log error
var LogError = New("error", 3)

//LogWarn log Warning
var LogWarn = New("warning", 4)

//LogInfo log Info
var LogInfo = New("information", 6)

//LogBuild log Debug
var LogBuild = New("debug", 7)

//Logger struct to logger
type Logger struct {
	*log.Logger
	Type string
}

//New create Logger
func New(ttype string, flag int) *Logger {
	return &Logger{
		Logger: log.New(os.Stderr, "", flag),
		Type:   ttype,
	}
}

//SetLogError set logs with ERROR level
func (logg *Logger) SetLogError(logger *log.Logger) {
	logg.Logger = logger
}

//Disable set logs with ERROR level
func (logg *Logger) Disable() {
	logg.Logger.SetOutput(ioutil.Discard)
}
