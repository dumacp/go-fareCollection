package app

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
)

// var count = 0

func (a *Actor) AwaitState(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", \"AwaitSate\", %v", ctx.Self().GetId(), ctx.Self())
	case *MsgReqAddress:
	case *usostransporte.MsgErrorDB:
		logstrans.LogError.Printf("usostransport DB is NOT ok")
		go func() {
			if a.pidUso != nil {
				time.Sleep(30 * time.Second)
				ctx.Request(a.pidUso, &usostransporte.MsgVerifyDB{})
			}
		}()
	case *usostransporte.MsgOkDB:
		logstrans.LogInfo.Printf("usostransport DB ok")
		if a.isReaderOk {
			a.fmachine.Event(eWait)
			a.behavior.Become(a.RunState)
		}
	case *messages.MsgSEOK:
		a.isReaderOk = true
		logstrans.LogInfo.Printf("SE OK")
		a.fmachine.Event(eWait)
		a.behavior.Become(a.RunState)
	}
}
