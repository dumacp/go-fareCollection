package graph

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/pubsub"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	topicGraph = "GRAPH/app"
)

type Actor struct {
	inputs int
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("%s - msg: %+v, type: %T", ctx.Self().GetId(), ctx.Message(), ctx.Message())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		// if err := pubsub.Init(); err != nil {
		// 	logs.LogError.Println(err)
		time.Sleep(3 * time.Second)
		// 	panic(err)
		// }
	case *MsgWaitTag:
		screen1 := &Screen{
			ID:  1,
			Msg: []string{msg.Message, msg.Ruta},
			// Msg: []string{"presente medio\r\nde pago", msg.Ruta},
		}
		sendMsg1, err := funScreen(screen1)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg1)

		time.Sleep(100 * time.Millisecond)
		count1 := new(Countinputs)
		count1.Count = a.inputs
		sendMsg3, err := funCounts(count1)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg3)
		time.Sleep(100 * time.Millisecond)
	case *MsgRef:
		// time.Sleep(100 * time.Millisecond)
		ref1 := new(ReferenceApp)
		ref1.Refproduct = msg.Device
		ref1.Appversion = fmt.Sprintf("V: %s", msg.Version)
		sendMsg2, err := funRef(ref1)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg2)
		time.Sleep(100 * time.Millisecond)
		count1 := new(Countinputs)
		count1.Count = a.inputs
		sendMsg3, err := funCounts(count1)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg3)

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
	case *MsgValidationQR:
		screen2 := &Screen{
			ID:  2,
			Msg: []string{`Ticket válido`, msg.Value},
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
	case *MsgError:
		screen3 := &Screen{
			ID:  3,
			Msg: msg.Value,
		}
		sendMsg, err := funScreen(screen3)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	case *MsgOk:
		screen2 := &Screen{
			ID:  2,
			Msg: msg.Value,
		}
		sendMsg, err := funScreen(screen2)
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
	case *MsgQrValue:
		qrvalur := &Qrvalue{URL: msg.Value}
		sendMsg, err := funQR(qrvalur)
		if err != nil {
			logs.LogWarn.Println(err)
			break
		}
		pubsub.Publish(topicGraph, sendMsg)
	case *MsgCount:
		a.inputs = msg.Value
		// ref1 := new(ReferenceApp)
		// ref1.Appversion = "V: "
		// ref1.Refproduct = "OMV-Z7-1431"
		// ref1.Appversion = fmt.Sprintf("%s%.02f", "V: ", 1.0)
		// ref1.Count = msg.Value
		// sendMsg1, err := funRef(ref1)
		// if err != nil {
		// 	logs.LogWarn.Println(err)
		// 	break
		// }
		// pubsub.Publish(topicGraph, sendMsg1)
		// counts := &Countinputs{Count: msg.Value}
		// sendMsg2, err := funCounts(counts)
		// if err != nil {
		// 	logs.LogWarn.Println(err)
		// 	break
		// }
		// pubsub.Publish(topicGraph, sendMsg2)
	}
}
