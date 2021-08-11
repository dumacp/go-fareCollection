package qr

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/looplab/fsm"
)

type Actor struct {
	fmachine *fsm.FSM
	ctx      actor.Context
}

func NewActor() actor.Actor {
	a := &Actor{}
	a.fmachine = NewFSM(nil)
	go a.RunFSM()
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.fmachine.Event(eOpened)
	case *MsgNewCodeQR:
		ctx.Send(ctx.Parent(), msg)
	}
}
