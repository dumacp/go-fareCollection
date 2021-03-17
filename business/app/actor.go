package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/business/graph"
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
	inputs         int
	fmachine       *fsm.FSM
	lastTime       time.Time
	ctx            actor.Context
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
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
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			panic(err)
		}
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
	case *MsgTagRead:
		if err := func() error {
			// if a.actualTag == a.errorWriteTag {
			// 	//Commit Tag
			// 	return nil
			// }
			if err := ValidationTag(msg.Data); err != nil {
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
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
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
