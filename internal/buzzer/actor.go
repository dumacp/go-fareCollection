package buzzer

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
)

type Actor struct {
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in buzzerActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch ctx.Message().(type) {

	case *actor.Started:

		if err := BuzzerInit(); err != nil {
			logs.LogError.Println(err)
			time.Sleep(3 * time.Second)
			// panic(err)
		}
	case *MsgBuzzerGood:
		BuzzerPlayGOOD()
	case *MsgBuzzerBad:
		BuzzerPlayBAD()
	}
}
