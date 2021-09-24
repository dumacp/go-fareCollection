package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/dumacp/go-fareCollection/internal/buzzer"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/picto"
	"github.com/dumacp/go-fareCollection/internal/pubsub"
	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-fareCollection/pkg/services"
	"github.com/dumacp/go-logs/pkg/logs"
	semessages "github.com/dumacp/go-sesam/pkg/messages"
	"github.com/looplab/fsm"
)

type Actor struct {
	behavior     actor.Behavior
	deviceID     string
	version      string
	deviceSerial int
	// lastTag        uint64
	// lastTimeDetect time.Time
	// errorWriteTag  uint64
	// actualTag      uint64
	pidGraph  *actor.PID
	pidBuzzer *actor.PID
	pidPicto  *actor.PID
	pidQR     *actor.PID
	pidFare   *actor.PID
	pidList   *actor.PID
	pidGps    *actor.PID
	pidUso    *actor.PID
	pidParams *actor.PID
	pidSam    *actor.PID
	fmachine  *fsm.FSM
	lastTime  time.Time
	ctx       actor.Context
	paym      map[uint64]payment.Payment
	recharge  *recharge.Recharge
	// rawcard   map[string]string
	// updates        map[string]interface{}
	// chNewRand  chan int
	oldRand int
	newRand int
	inputs  int
	outputs int
	seq     int
	// itineraryMap   itinerary.ItineraryMap
	params          *parameters.Parameters
	listRestrictive map[string]*lists.WatchList
	timeout         int
	lastWriteError  uint64
	isReaderOk      bool
	quit            chan int
}

func NewActor(id string, version string) actor.Actor {
	a := &Actor{}
	a.behavior = make(actor.Behavior, 0)
	a.behavior.Become(a.InitState)
	a.deviceID = id
	a.version = version
	varSplit := strings.Split(id, "-")
	if len(varSplit) > 0 {
		a.deviceSerial, _ = strconv.Atoi(varSplit[len(varSplit)-1])
	}
	a.newFSM(nil)
	go a.RunFSM()
	return a
}

// var count = 0

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	logs.LogBuild.Printf("Message arrived in appActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Stopping:
		logs.LogWarn.Printf("\"%s\" - Stopped actor, reason -> %v", ctx.Self(), msg)
		ctx.Send(ctx.Self(), &MsgStop{})
	case *MsgStop:
		a.writeerrorverify()
		if a.pidParams != nil {
			ctx.Send(a.pidParams, &parameters.AppParameters{
				Seq:     uint(a.seq),
				Inputs:  a.inputs,
				Outputs: a.outputs,
			})
		}
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", %v", ctx.Self().GetId(), ctx.Self())
		if err := func() error {
			propsBuzzer := actor.PropsFromProducer(buzzer.NewActor)
			pidBuzzer, err := ctx.SpawnNamed(propsBuzzer, "buzzer-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidBuzzer = pidBuzzer
			propsPicto := actor.PropsFromProducer(picto.NewActor)
			pidPicto, err := ctx.SpawnNamed(propsPicto, "picto-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidPicto = pidPicto
			propsQR := actor.PropsFromProducer(qr.NewActor)
			pidQR, err := ctx.SpawnNamed(propsQR, "qr-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidQR = pidQR

			pubsub.Subscribe(services.TopicAddress, ctx.Self(),
				func(msg []byte) interface{} {
					return &MsgReqAddress{Addr: string(msg)}
				})
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			time.Sleep(3 * time.Second)
			panic(err)
		}
		// a.fmachine.Event(eStarted)
		// ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)
	case *MsgWriteAppParams:
		if a.pidParams != nil {
			ctx.Send(a.pidParams, &parameters.AppParameters{
				Seq:     uint(a.seq),
				Inputs:  a.inputs,
				Outputs: a.outputs,
			})
		}
	case *parameters.ConfigParameters:
		if a.pidParams != nil {
			ctx.Send(a.pidParams, msg)
		}
		a.fmachine.Event(eOk, []string{"configuración", "datos actualizados"})
	case *parameters.MsgParameters:
		logstrans.LogInfo.Printf("params: %+v", msg.Data)
		if ctx.Sender() != nil {
			a.pidParams = ctx.Sender()
		}

		a.params = msg.Data
		a.timeout = a.params.Timeout
		if a.params.Seq >= uint(a.seq) {
			a.seq = int(a.params.Seq)
		}
		if a.params.Inputs >= a.inputs {
			a.inputs = int(a.params.Inputs)
		}
		if a.params.Outputs >= a.outputs {
			a.outputs = int(a.params.Outputs)
		}
		if a.params.DevSerial > 0 {
			a.deviceSerial = a.params.DevSerial
		}
		if a.pidList != nil {
			for _, list := range a.params.RestrictiveList {
				ctx.Request(a.pidList, &lists.MsgWatchList{ID: list})
			}
		}

	case *lists.WatchList:
		if a.listRestrictive == nil {
			a.listRestrictive = make(map[string]*lists.WatchList)
		}
		a.listRestrictive[msg.ID] = msg
		logstrans.LogInfo.Printf("watch over lists: id: %v, type: %v, version: %v",
			msg.ID, msg.PaymentMediumType, msg.Version)
	case *MsgReqAddress:

	case *messages.RegisterGraphActor:
		a.pidGraph = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterFareActor:
		a.pidFare = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterListActor:
		a.pidList = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterGPSActor:
		a.pidGps = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterUSOActor:
		a.pidUso = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterSAMActor:
		a.pidSam = actor.NewPID(msg.Addr, msg.Id)
		ctx.Request(a.pidSam, &messages.MsgSERequestStatus{})
	case *qr.MsgNewCodeQR:
		if err := func() error {
			divInput := msg.Value[0:4]
			revDiv := make([]byte, 0)
			for i := range divInput {
				revDiv = append(revDiv, divInput[len(divInput)-1-i])
			}
			iv := make([]byte, 0)
			iv = append(iv, divInput...)
			iv = append(iv, revDiv...)
			iv = append(iv, divInput...)
			iv = append(iv, revDiv...)
			keySlot := 51
			if a.params != nil && a.params.KeyQr > 0 {
				keySlot = a.params.KeyQr
			}
			mdec := &semessages.MsgDecryptRequest{
				Data:     msg.Value[4:],
				DevInput: divInput[0:4],
				IV:       iv,
				//TODO: How to get ?
				KeySlot: keySlot,
			}
			// logs.LogBuild.Printf("QR crypt: [% X], len: %d; [% X], len: %d",
			// 	mdec.Data, len(mdec.Data), mdec.DevInput, len(mdec.DevInput))

			var data []byte
			samuid := ""

			res, err := ctx.RequestFuture(a.pidSam, mdec, time.Millisecond*600).Result()
			if err != nil {
				return fmt.Errorf("get decrypt sam err: %w", err)
			}
			switch v := res.(type) {
			case *semessages.MsgDecryptResponse:
				if v.Plain == nil {
					return errors.New("QR decrypt data is empty")
				}
				data = v.Plain
				samuid = v.SamUID
				logs.LogBuild.Printf("QR decrypt: %s, [%X]", data, data)
			}

			if ctx.Sender() != nil {
				ctx.Respond(&qr.MsgResponseCodeQR{
					Value:  data,
					SamUid: samuid,
				})
			}
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("QR error: %s", err)
		}
	}
	a.behavior.Receive(ctx)
}
