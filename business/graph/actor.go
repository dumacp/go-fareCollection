package graph

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/crosscutting/comm/pubsub"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
)

const (
	topicGraph = "GRAPH/app"
)

type Actor struct{}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if err := pubsub.Init(); err != nil {
			logs.LogError.Println(err)
			time.Sleep(3 * time.Second)
			panic(err)
		}
	case *MsgWaitTag:
		screen1 := &Screen{
			ID:  1,
			Msg: []string{"presente medio\r\nde pago", `SIITP`},
		}
		sendMsg, err := funScreen(screen1)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	case *MsgValidationTag:
		screen2 := &Screen{
			ID:  2,
			Msg: []string{`Saldo disponible`, msg.Value},
		}
		sendMsg, err := funScreen(screen2)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	case *MsgBalanceError:
		screen3 := &Screen{
			ID:  3,
			Msg: []string{`Saldo insuficiente`, msg.Value},
		}
		sendMsg, err := funScreen(screen3)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	case *MsgWriteError:
		screen3 := &Screen{
			ID:  3,
			Msg: []string{`Error de escritura`, `intentelo de nuevo`},
		}
		sendMsg, err := funScreen(screen3)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	}
}
