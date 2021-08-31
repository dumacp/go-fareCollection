package logstrans

import (
	"io/ioutil"
	"log"
	"os"
)

//LogError log error
var LogError = New("error", log.Ldate|log.Ltime)

//LogWarn log Warning
var LogWarn = New("warning", log.Ldate|log.Ltime)

//LogInfo log Info
var LogInfo = New("information", log.Ldate|log.Ltime)

//LogBuild log Debug
var LogBuild = New("debug", log.Ldate|log.Ltime)

//logTransaction log transaction
var LogTransInfo = New("information", log.Ldate|log.Ltime)

//logTransaction log transaction
var LogTransWarn = New("warning", log.Ldate|log.Ltime)

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
