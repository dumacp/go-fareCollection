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
	"github.com/dumacp/go-fareCollection/internal/gps"
	"github.com/dumacp/go-fareCollection/internal/itinerary"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/pubsub"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/smartcard/multiiso"
)

var debug bool
var logstd bool
var id string

func init() {
	flag.BoolVar(&debug, "debug", false, "debug?")
	flag.BoolVar(&logstd, "logStd", false, "logs in stderr?")
	flag.StringVar(&id, "id", "OMZV7-0001", "device ID")
}

func main() {
	flag.Parse()
	initLogs(debug, logstd)

	logs.LogBuild.Println("debug log")

	ctx := actor.NewActorSystem().Root

	pubsub.Init(ctx)

	propsGps := actor.PropsFromProducer(gps.NewActor)

	pidGps, err := ctx.SpawnNamed(propsGps, "gps-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	propsFare := actor.PropsFromProducer(fare.NewActor)
	pidFare, err := ctx.SpawnNamed(propsFare, "fare-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	itiActor := itinerary.NewActor()
	propsIti := actor.PropsFromFunc(itiActor.Receive)

	pidIti, err := ctx.SpawnNamed(propsIti, "iti-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	paramActor := parameters.NewActor(id)
	propsParam := actor.PropsFromFunc(paramActor.Receive)

	pidParam, err := ctx.SpawnNamed(propsParam, "param-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	propsList := actor.PropsFromProducer(lists.NewActor)

	pidList, err := ctx.SpawnNamed(propsList, "list-actor")
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

	ctx.Send(pidApp, &messages.RegisterGPSActor{Addr: pidGps.Address, Id: pidGps.Id})
	ctx.Send(pidApp, &messages.RegisterFareActor{Addr: pidFare.Address, Id: pidFare.Id})
	ctx.Send(pidApp, &messages.RegisterListActor{Addr: pidList.Address, Id: pidList.Id})
	ctx.RequestWithCustomSender(pidIti, &itinerary.MsgSubscribe{}, pidFare)
	// ctx.RequestWithCustomSender(pidIti, &itinerary.MsgSubscribe{}, pidApp)
	ctx.RequestWithCustomSender(pidParam, &parameters.MsgSubscribe{}, pidApp)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)

	for range finish {
		log.Print("Finish")
		return
	}
}
