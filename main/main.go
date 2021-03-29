package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AsynkronIT/protoactor-go/actor"
	appreader "github.com/dumacp/go-appliance-contactless/business/app"
	"github.com/dumacp/go-fareCollection/business/app"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
)

var debug bool
var logstd bool

func init() {
	flag.BoolVar(&debug, "debug", false, "debug?")
	flag.BoolVar(&logstd, "logStd", false, "logs in stderr?")
}

func main() {
	flag.Parse()
	initLogs(debug, logstd)

	logs.LogBuild.Println("debug log")

	ctx := actor.EmptyRootContext

	appActor := app.NewActor()
	propsApp := actor.PropsFromFunc(appActor.Receive)

	pidApp, err := ctx.SpawnNamed(propsApp, "app-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	readerActor := appreader.NewActor(pidApp)
	propsReader := actor.PropsFromFunc(readerActor.Receive)

	_, err = ctx.SpawnNamed(propsReader, "reader-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)

	for {
		select {
		case <-finish:
			log.Print("Finish")
			return
		}
	}

}
