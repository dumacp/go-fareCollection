package qr

import (
	"github.com/AsynkronIT/protoactor-go/actor"
)

type Actor struct {
}

func NewActor() actor.Actor {

	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {

	case *actor.Started:

	}
}
