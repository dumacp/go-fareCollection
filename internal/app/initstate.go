package app

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

// var count = 0

func (a *Actor) InitState(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Stopping:
		logs.LogWarn.Printf("\"%s\" - Stopped actor, reason -> %v", ctx.Self(), msg)
		ctx.Send(ctx.Self(), &MsgWriteAppParams{})
		ctx.Send(ctx.Self(), &MsgWriteErrorVerify{})
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", \"RunSate\", %v", ctx.Self().GetId(), ctx.Self())
	case *parameters.MsgParameters:
		if a.isReaderOk {
			a.fmachine.Event(eWait)
			a.behavior.Become(a.RunState)
		}
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "app params OK ...",
				Percent: 85,
			}
			ctx.Send(a.pidGraph, screen0)
			time.Sleep(1 * time.Second)
			a.fmachine.Event(eWait)
		}
	case *MsgReqAddress:
	case *messages.RegisterGraphActor:
		a.pidGraph = actor.NewPID(msg.Addr, msg.Id)

		screen0 := &graph.Loading{
			ID:      0,
			Msg:     "picto OK ...",
			Percent: 90,
		}
		ctx.Send(a.pidGraph, screen0)
	case *parameters.MsgStatus:
		if !msg.State {
			if a.pidGraph != nil {
				screen0 := &graph.Loading{
					ID:      0,
					Msg:     "params is NOT ok ...",
					Percent: -1,
				}
				ctx.Send(a.pidGraph, screen0)
			}
			if a.pidParams != nil {
				go func() {
					time.Sleep(10 * time.Second)
					ctx.Send(a.pidParams, &parameters.MsgRequestStatus{})
				}()
			}
		} else {
			if a.isReaderOk {
				a.fmachine.Event(eWait)
				a.behavior.Become(a.RunState)
			}
		}
	case *usostransporte.MsgOkDB:
		a.disableApp = false
		logstrans.LogInfo.Printf("usostransport DB ok")
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "app db OK ...",
				Percent: -1,
			}
			ctx.Send(a.pidGraph, screen0)
		}
	case *messages.MsgSEError:
		a.isReaderOk = false
		logstrans.LogError.Printf("--- SE error, err: %s", msg.Error)
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "reader is NOT ok ...",
				Percent: -1,
			}
			ctx.Send(a.pidGraph, screen0)
		}
	case *messages.MsgSEOK:
		a.isReaderOk = true
		logstrans.LogInfo.Printf("SE OK")
		screen0 := &graph.Loading{
			ID:      0,
			Msg:     "reader OK ...",
			Percent: 91,
		}
		if a.params != nil {
			if a.pidGraph != nil {
				ctx.Send(a.pidGraph, screen0)
				time.Sleep(1 * time.Second)
			}
		} else {
			if a.pidParams != nil {
				ctx.Send(a.pidParams, &parameters.MsgRequestStatus{})
			}
		}
		if a.params != nil {
			a.fmachine.Event(eWait)
			a.behavior.Become(a.RunState)
		}
	}
}
