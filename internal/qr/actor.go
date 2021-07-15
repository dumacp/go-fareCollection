package qr

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/looplab/fsm"
)

type Actor struct {
	ch           chan int
	chQuitSerial chan int
	fmachine     *fsm.FSM
	ctx          actor.Context
}

func NewActor() actor.Actor {

	a := &Actor{}
	a.fmachine = NewFSM(nil)
	a.ch = make(chan int)
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if a.chQuitSerial != nil {
			select {
			case _, ok := <-a.chQuitSerial:
				if ok {
					close(a.chQuitSerial)
				}
			default:
				close(a.chQuitSerial)
			}
		}
		a.chQuitSerial = make(chan int)
		go RunFSM(ctx, a.chQuitSerial, a.fmachine)
		a.fmachine.Event(eOpened)
	case *MsgNewCodeQR:
		ctx.Send(ctx.Parent(), msg)
	}
}
