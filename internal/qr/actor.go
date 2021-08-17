package qr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

type Actor struct {
	fmachine *fsm.FSM
	ctx      actor.Context
	// actualRand int
	// lastRand   int
	quit   chan int
	chRand chan int
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
		a.chRand = make(chan int)
		go tickQR(ctx, a.quit, a.chRand)
	case *MsgNewCodeQR:
		ctx.Request(ctx.Parent(), msg)
	case *MsgNewRand:
		// a.lastRand = a.actualRand
		// a.actualRand = msg.Value

		if ctx.Parent() != nil {
			ctx.Send(ctx.Parent(), msg)
		}
	case *messages.MsgWritePayment:
		select {
		case a.chRand <- 1:
		case <-time.After(100 * time.Millisecond):
		}
		if ctx.Sender() != nil {
			ctx.Send(ctx.Parent(), &messages.MsgWritePaymentResponse{
				Uid:  msg.Uid,
				Type: msg.Type,
				Raw:  make(map[string]string),
			})
		}
	case *MsgResponseCodeQR:
		if err := func() error {
			data := msg.Value
			mess := struct {
				Type  string          `json:"t"`
				Value json.RawMessage `json:"v"`
			}{}
			data = bytes.TrimRight(data, "\x00")
			if err := json.Unmarshal(data, &mess); err != nil {
				return fmt.Errorf("QR error: %w", err)
			}
			logs.LogBuild.Printf("NewQR: %s", msg.Value)
			switch strings.ToUpper(mess.Type) {

			case "DCIT":
				sendMsg := new(parameters.ConfigParameters)
				if err := json.Unmarshal(mess.Value, sendMsg); err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				if ctx.Parent() != nil {
					ctx.Send(ctx.Parent(), sendMsg)
				}
			case "EQPM":
				var mapp map[string]interface{}
				if err := json.Unmarshal(mess.Value, &mapp); err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				raw := make(map[string]string)
				data := make(map[string]*messages.Value)
				for k, v := range mapp {
					switch value := v.(type) {
					case float64:
						raw[k] = fmt.Sprintf("%d", int32(value))
						data[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
					case int:
						raw[k] = fmt.Sprintf("%d", value)
						data[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
					case uint:
						raw[k] = fmt.Sprintf("%d", value)
						data[k] = &messages.Value{Data: &messages.Value_UintValue{UintValue: uint32(value)}}
					case int64:
						raw[k] = fmt.Sprintf("%d", value)
						data[k] = &messages.Value{Data: &messages.Value_Int64Value{Int64Value: int64(value)}}
					case uint64:
						raw[k] = fmt.Sprintf("%d", value)
						data[k] = &messages.Value{Data: &messages.Value_Uint64Value{Uint64Value: uint64(value)}}
					case string:
						raw[k] = value
						data[k] = &messages.Value{Data: &messages.Value_StringValue{StringValue: value}}
					case []byte:
						raw[k] = fmt.Sprintf("%s", value)
						data[k] = &messages.Value{Data: &messages.Value_BytesValue{BytesValue: value}}
					}
				}
				if ctx.Parent() != nil {
					ctx.Request(ctx.Parent(), &messages.MsgPayment{
						Type: "ENDUSER_QR",
						Data: data,
						Raw:  raw,
					})
				}

				// res := new(QrCode)
				// if err := json.Unmarshal(newv, res); err != nil {
				// 	logs.LogError.Printf("QR error: %s", err)
				// 	break
				// }
				// pin, err := strconv.Atoi(res.Pin)
				// if err != nil {
				// 	logs.LogError.Printf("QR error: %s", err)
				// 	break
				// }
				// if pin != a.lastRand && pin != a.actualRand {
				// 	a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
				// 	a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
				// 	//TODO: cahngeeee!!!
				// 	go func() {
				// 		time.Sleep(2 * time.Second)
				// 		a.ctx.Send(a.pidPicto, &picto.MsgPictoOFF{})
				// 	}()
				// 	logs.LogError.Printf("QR error: PIN is invalid")
				// 	break
				// }
				// a.fmachine.Event(eQRValidated, fmt.Sprintf("%d", res.TransactionID))
				// // ctx.Send(a.pidGraph, &graph.MsgValidationQR{Value: fmt.Sprintf("%d", res.TransactionID)})

				// select {
				// case a.chNewRand <- 1:
				// case <-time.After(100 * time.Millisecond):
				// }
				// a.lastRand = a.actualRand
				// a.actualRand = -1

				// // go func() {
				// SendUsoQR(int(res.TransactionID), []float64{0, 0}, time.Now())
				// if err != nil {
				// 	return fmt.Printf("POST error: %s", err)
				// }
				// logs.LogInfo.Printf("response platform: %s", response)
			}
			return nil
		}(); err != nil {
			logs.LogError.Printf(err.Error())
		}
	}
}
