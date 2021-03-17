package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	appreader "github.com/dumacp/go-appliance-contactless/business/app"
	"github.com/dumacp/go-fareCollection/business/buzzer"
	"github.com/dumacp/go-fareCollection/business/graph"
	"github.com/dumacp/go-fareCollection/business/picto"
	"github.com/dumacp/go-fareCollection/business/qr"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
	"github.com/looplab/fsm"
)

type Actor struct {
	lastTag        uint64
	lastTimeDetect time.Time
	errorWriteTag  uint64
	actualTag      uint64
	pidGraph       *actor.PID
	pidBuzzer      *actor.PID
	pidPicto       *actor.PID
	pidReader      *actor.PID
	pidQR          *actor.PID
	inputs         int
	fmachine       *fsm.FSM
	lastTime       time.Time
	ctx            actor.Context
	mcard          map[string]interface{}
}

func NewActor() actor.Actor {
	a := &Actor{}
	a.newFSM(nil)
	go a.RunFSM()
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if err := func() error {
			propsGrpah := actor.PropsFromProducer(graph.NewActor)
			pidGrpah, err := ctx.SpawnNamed(propsGrpah, "graph-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidGraph = pidGrpah
			propsBuzzer := actor.PropsFromProducer(buzzer.NewActor)
			pidBuzzer, err := ctx.SpawnNamed(propsBuzzer, "buzzer-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidBuzzer = pidBuzzer
			propsPicto := actor.PropsFromProducer(picto.NewActor)
			pidPicto, err := ctx.SpawnNamed(propsPicto, "buzzer-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidPicto = pidPicto
			propsQR := actor.PropsFromProducer(qr.NewActor)
			pidQR, err := ctx.SpawnNamed(propsQR, "buzzer-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidQR = pidQR
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			time.Sleep(3 * time.Second)
			panic(err)
		}
		a.fmachine.Event(eStarted)
		ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
	case *MsgTagDetected:
		if err := func() error {
			if a.actualTag == a.lastTag {
				if a.lastTimeDetect.Before(time.Now().Add(-5 * time.Second)) {
					return nil
				}
			}
			//read card
			ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
			a.lastTag = msg.UID
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *appreader.MsgCardRead:
		logs.LogBuild.Printf("tag read: %v", msg.Map)
		if err := func() error {
			// if a.actualTag == a.errorWriteTag {
			// 	//Commit Tag
			// 	return nil
			// }
			if a.pidReader == nil {
				a.pidReader = ctx.Sender()
			}
			a.mcard = make(map[string]interface{})
			for k, v := range msg.Map {
				a.mcard[k] = v
			}
			v, err := ValidationTag(a.mcard)
			if err != nil {
				logs.LogBuild.Println(err)
				if errors.Is(err, ErrorBalance) {
					//Send Msg Error Balance

					if balanceErr, ok := err.(*ErrorBalanceValue); ok {
						ctx.Send(a.pidGraph, &graph.MsgBalanceError{Value: fmt.Sprintf("%.02f", balanceErr.Balance)})
					} else {
						ctx.Send(a.pidGraph, &graph.MsgBalanceError{Value: ""})
					}
				}
				time.Sleep(3 * time.Second)
				return err
			}
			ctx.Send(ctx.Sender(), &appreader.MsgWriteCard{UID: msg.UID, Updates: v})
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *appreader.MsgCardWritten:
		ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: fmt.Sprintf("$%.02f", float32(a.mcard["newSaldo"].(int32)))})
		a.fmachine.Event(eCardValidated)
	case *MsgQRRead:
		if err := func() error {
			if a.actualTag == a.errorWriteTag {
				//Commit Tag
				return nil
			}
			if err := ValidationQr(msg.Data); err != nil {
				if errors.Is(err, ErrorQR) {
					//Send Msg Error Balance
					defer ctx.Send(nil, nil)
				}
				time.Sleep(3 * time.Second)
				return err
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgTagWriteError:
		if err := func() error {
			//Send Msg Write error
			defer ctx.Send(nil, nil)
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	}
}
