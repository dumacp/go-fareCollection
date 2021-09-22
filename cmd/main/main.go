package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/mailbox"
	"github.com/AsynkronIT/protoactor-go/remote"
	appreader "github.com/dumacp/go-appliance-contactless/pkg/app"
	"github.com/dumacp/go-fareCollection/internal/app"
	"github.com/dumacp/go-fareCollection/internal/fare"
	"github.com/dumacp/go-fareCollection/internal/gps"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/itinerary"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/pubsub"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/smartcard"
	"github.com/dumacp/smartcard/multiiso"
	"github.com/google/uuid"
)

const (
	version = "1.0.1"
)

var verbose int
var logstd bool
var id string
var serial string
var baud int
var dirlogs string
var prefixlogs string

func init() {
	flag.IntVar(&verbose, "verbose", 0, "level log\n\t0: Error\n\t1: Warning\n\t2: Info\n\t3: Debug")
	flag.BoolVar(&logstd, "logStd", false, "logs in stderr?")
	flag.StringVar(&id, "id", "", "device ID")
	flag.StringVar(&serial, "serial", "/dev/ttymxc4", "device path")
	flag.IntVar(&baud, "baud", 460800, "device baud speed")
	flag.StringVar(&dirlogs, "dirlogs", "/SD/logs/", "dir path to logs")
	flag.StringVar(&prefixlogs, "prefixlogs", "log", "prefix to filename")
}

func main() {
	flag.Parse()
	initLogs(verbose, logstd)
	initLogsTransactional(dirlogs, prefixlogs, verbose, logstd)

	fmt.Printf("url: %s\n", utils.Url)

	logs.LogBuild.Println("debug log")

	sys := actor.NewActorSystem()

	portlocal := 8009
	for {
		portlocal++

		socket := fmt.Sprintf("127.0.0.1:%d", portlocal)
		testConn, err := net.DialTimeout("tcp", socket, 1*time.Second)
		if err != nil {
			break
		}
		testConn.Close()
		logs.LogWarn.Printf("socket busy -> \"%s\"", socket)
		time.Sleep(1 * time.Second)
	}

	rconfig := remote.Configure("127.0.0.1", portlocal).WithServerOptions()
	r := remote.NewRemote(sys, rconfig)
	r.Start()

	// decider := func(reason interface{}) actor.Directive {
	// 	fmt.Println("handling failure for child")
	// 	return actor.StopDirective
	// }
	// supervisor := actor.NewOneForOneStrategy(10, 1000, decider)

	// ctx := sys.Root.WithGuardian(supervisor)
	ctx := sys.Root

	pubsub.Init(ctx)

	propsGrpah := actor.PropsFromProducer(graph.NewActor)
	pidGrpah, err := ctx.SpawnNamed(propsGrpah, "graph-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	screen0 := &graph.Loading{
		ID:      0,
		Msg:     "graph OK ...",
		Percent: 5,
	}

	ctx.Send(pidGrpah, screen0)

	propsGps := actor.PropsFromProducer(gps.NewActor)

	pidGps, err := ctx.SpawnNamed(propsGps, "gps-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(1 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "gps OK ...",
		Percent: 10,
	}

	ctx.Send(pidGrpah, screen0)

	propsFare := actor.PropsFromProducer(fare.NewActor).WithMailbox(mailbox.UnboundedPriority())
	pidFare, err := ctx.SpawnNamed(propsFare, "fare-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(1 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "fare policies OK ...",
		Percent: 20,
	}
	ctx.Send(pidGrpah, screen0)

	itiActor := itinerary.NewActor()
	propsIti := actor.PropsFromFunc(itiActor.Receive)

	pidIti, err := ctx.SpawnNamed(propsIti, "iti-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(2 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "itineraries map OK ...",
		Percent: 40,
	}
	ctx.Send(pidGrpah, screen0)

	//TODO: get id from hostname?

	if len(id) > 0 {
		utils.SetHostname(id)
	}
	id = utils.Hostname()
	logs.LogBuild.Printf("ID: %s", id)
	varSplit := strings.Split(id, "-")
	if len(varSplit) < 1 {
		id = "0001"
	}
	idint, _ := strconv.Atoi(varSplit[len(varSplit)-1])

	idbytes := make([]byte, 8)

	binary.LittleEndian.PutUint64(idbytes, uint64(idint))
	logs.LogBuild.Printf("nodeID: [% X]", idbytes[:6])
	uuid.SetNodeID(idbytes[:6])

	paramActor := parameters.NewActor(id)
	propsParam := actor.PropsFromFunc(paramActor.Receive)

	pidParam, err := ctx.SpawnNamed(propsParam, "param-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(2 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "parameters OK ...",
		Percent: 60,
	}
	ctx.Send(pidGrpah, screen0)

	propsList := actor.PropsFromProducer(lists.NewActor).WithMailbox(mailbox.UnboundedPriority())

	pidList, err := ctx.SpawnNamed(propsList, "list-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(1 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "lists OK ...",
		Percent: 70,
	}
	ctx.Send(pidGrpah, screen0)

	propsUsos := actor.PropsFromProducer(usostransporte.NewActor)

	pidUsos, err := ctx.SpawnNamed(propsUsos, "usos-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(2 * time.Second)

	screen0 = &graph.Loading{
		ID:      0,
		Msg:     "usos databases OK ...",
		Percent: 80,
	}
	ctx.Send(pidGrpah, screen0)

	appActor := app.NewActor(id, version)
	propsApp := actor.PropsFromFunc(appActor.Receive)

	pidApp, err := ctx.SpawnNamed(propsApp, "app-actor")
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	// init contactless reader
	var dev *multiiso.Device
	funcReader := func() smartcard.IReader {
		exec.Command("/bin/sh", "-c", "echo 0 > /sys/class/leds/enable-reader/brightness").Run()
		time.Sleep(1 * time.Second)
		if res, err := exec.Command("/bin/sh", "-c", "echo 1 > /sys/class/leds/enable-reader/brightness").CombinedOutput(); err != nil {
			logs.LogError.Printf("%s, err: %s", res, err)
			logstrans.LogError.Printf("%s, err: %s", res, err)
		}
		time.Sleep(1 * time.Second)

		dev, err = multiiso.NewDevice(serial, baud, time.Millisecond*600)
		if err != nil {
			logs.LogError.Fatalln(err)
		}
		reader := multiiso.NewReader(dev, "multiiso", 1)
		return reader
	}

	// TEST
	// go func() {
	// 	time.Sleep(20 * time.Second)
	// 	dev.Close()
	// 	time.Sleep(30 * time.Second)
	// 	dev.Close()
	// }()

	readerActor, err := appreader.NewActor(ctx, funcReader)
	if err != nil {
		logs.LogError.Fatalln(err)
	}

	time.Sleep(6 * time.Second)

	readerActor.Subscribe(pidApp)

	ctx.Send(pidApp, &messages.RegisterGraphActor{Addr: pidGrpah.Address, Id: pidGrpah.Id})
	ctx.Send(pidApp, &messages.RegisterGPSActor{Addr: pidGps.Address, Id: pidGps.Id})
	ctx.Send(pidApp, &messages.RegisterFareActor{Addr: pidFare.Address, Id: pidFare.Id})
	ctx.Send(pidApp, &messages.RegisterListActor{Addr: pidList.Address, Id: pidList.Id})
	ctx.Send(pidApp, &messages.RegisterUSOActor{Addr: pidUsos.Address, Id: pidUsos.Id})
	//TODO: change actor sam
	ctx.Send(pidApp, &messages.RegisterSAMActor{Addr: readerActor.PID().Address, Id: readerActor.PID().Id})
	//TODO: first param
	ctx.RequestWithCustomSender(pidIti, &itinerary.MsgSubscribe{}, pidParam)
	ctx.RequestWithCustomSender(pidIti, &itinerary.MsgSubscribe{}, pidFare)
	time.Sleep(1 * time.Second)
	ctx.RequestWithCustomSender(pidParam, &parameters.MsgSubscribe{}, pidApp)

	// ctx.RequestWithCustomSender(pidIti, &itinerary.MsgSubscribe{}, pidApp)

	finish := make(chan os.Signal, 1)
	signal.Notify(finish, syscall.SIGINT)
	signal.Notify(finish, syscall.SIGTERM)

	for range finish {
		ctx.Poison(pidApp)
		time.Sleep(1 * time.Second)
		log.Print("Finish")
		return
	}
}
