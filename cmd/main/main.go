package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	appreader "github.com/dumacp/go-appliance-contactless/pkg/app"
	"github.com/dumacp/go-fareCollection/internal/app"
	"github.com/dumacp/go-fareCollection/internal/fare"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/smartcard/multiiso"
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

	ctx := actor.NewActorSystem().Root

	propsFare := actor.PropsFromProducer(fare.NewActor)
	pidFare, err := ctx.SpawnNamed(propsFare, "fare-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	appActor := app.NewActor()
	propsApp := actor.PropsFromFunc(appActor.Receive)

	pidApp, err := ctx.SpawnNamed(propsApp, "app-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	// init contactless reader
	dev, err := multiiso.NewDevice("/dev/ttyUSB0", 115200, time.Millisecond*600)
	if err != nil {
		logs.LogError.Fatalln(err)
	}
	reader := multiiso.NewReader(dev, "multiiso", 1)

	readerActor, err := appreader.NewActor(ctx, reader)
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	readerActor.Subscribe(pidApp)

	ctx.Send(pidFare, &messages.RegisterFareActor{Addr: pidFare.Address, Id: pidFare.Id})

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)

	for range finish {
		log.Print("Finish")
		return
	}
}
