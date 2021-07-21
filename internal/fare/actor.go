package fare

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/logs"
)

type Actor struct {
	farePolicies map[int]*FareNode
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in readerActor: %+v, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		ctx.Send(ctx.Self(), &MsgGetFarePolicies{})
	case *MsgGetFarePolicies:
		//TODO:
		//Get Fare Policies from platform
	case *MsgGetFare:
		calculate(msg.LastFarePolicies, a.farePolicies)
	}
}
