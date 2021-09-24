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
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", \"InitSate\", %v", ctx.Self().GetId(), ctx.Self())
	case *parameters.MsgParameters:
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "app params OK ...",
				Percent: 90,
			}
			ctx.Send(a.pidGraph, screen0)
			time.Sleep(1 * time.Second)
		}
		switch {
		case a.isReaderOk && a.params != nil && a.params.PaymentItinerary > 0:
			a.fmachine.Event(eStarted)
			a.behavior.Become(a.RunState)
		case a.params != nil && a.params.PaymentItinerary == 0:
			if a.pidGraph != nil {
				screen0 := &graph.Loading{
					ID:      0,
					Msg:     "itinerary is NOT ok",
					Percent: -1,
				}
				ctx.Send(a.pidGraph, screen0)
			}
		}
	case *messages.RegisterGraphActor:
		screen0 := &graph.Loading{
			ID:      0,
			Msg:     "picto OK ...",
			Percent: 90,
		}
		ctx.Send(a.pidGraph, screen0)
	case *MsgReqAddress:

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
			switch {
			case a.isReaderOk && a.params != nil && a.params.PaymentItinerary > 0:
				a.fmachine.Event(eStarted)
				a.behavior.Become(a.RunState)
			case a.params != nil && a.params.PaymentItinerary == 0:
				if a.pidGraph != nil {
					screen0 := &graph.Loading{
						ID:      0,
						Msg:     "itinerary is NOT ok",
						Percent: -1,
					}
					ctx.Send(a.pidGraph, screen0)
				}
			}
		}
	case *usostransporte.MsgOkDB:
		logstrans.LogInfo.Printf("usostransport DB ok")
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "app db OK ...",
				Percent: -1,
			}
			time.Sleep(1 * time.Second)
			ctx.Send(a.pidGraph, screen0)
		}
		if a.params != nil && a.isReaderOk {
			a.behavior.Become(a.RunState)
			a.fmachine.Event(eStarted)
		}
	case *usostransporte.MsgErrorDB:
		logstrans.LogInfo.Printf("usostransport DB is NOT ok")
		if a.pidGraph != nil {
			screen0 := &graph.Loading{
				ID:      0,
				Msg:     "app db is NOT ok ...",
				Percent: -1,
			}
			ctx.Send(a.pidGraph, screen0)
		}
		go func() {
			if a.pidUso != nil {
				time.Sleep(20 * time.Second)
				ctx.Send(a.pidUso, &usostransporte.MsgVerifyDB{})
			}
		}()
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
		go func() {
			time.Sleep(5 * time.Second)
			if a.pidSam != nil {
				ctx.Request(a.pidSam, &messages.MsgSERequestStatus{})
			}
		}()
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
			go func() {
				if a.pidParams != nil {
					time.Sleep(30 * time.Second)
					ctx.Send(a.pidParams, &parameters.MsgRequestStatus{})
				}
			}()
		}
		switch {
		case a.params != nil && a.params.PaymentItinerary > 0:
			a.fmachine.Event(eStarted)
			a.behavior.Become(a.RunState)
		case a.params != nil && a.params.PaymentItinerary == 0:
			if a.pidGraph != nil {
				screen0 := &graph.Loading{
					ID:      0,
					Msg:     "itinerary is NOT ok",
					Percent: -1,
				}
				ctx.Send(a.pidGraph, screen0)
			}
		}
	}
}
