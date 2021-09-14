package qr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

type Actor struct {
	fmachine *fsm.FSM
	ctx      actor.Context
	samuid   string
	chRand   chan int
	quit     chan int
}

func NewActor() actor.Actor {
	a := &Actor{}
	a.fmachine = NewFSM(nil)

	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	logs.LogBuild.Printf("Message arrived in qrActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.fmachine.Event(eOpened)
		a.quit = make(chan int)
		go a.RunFSM(a.quit)
		a.chRand = make(chan int)
		go tickQR(ctx, a.chRand)
	case *actor.Stopping:
		close(a.chRand)
		close(a.quit)
	case *MsgNewCodeQR:
		ctx.Request(ctx.Parent(), msg)
	case *MsgNewRand:
		if ctx.Parent() != nil {
			ctx.Send(ctx.Parent(), msg)
		}
	case *messages.MsgDetectPayment:
	case *messages.MsgWritePayment:
		switch msg.Type {
		case EQPM:
			select {
			case a.chRand <- 1:
			case <-time.After(100 * time.Millisecond):
			}
		}
		if ctx.Sender() != nil {
			ctx.Send(ctx.Parent(), &messages.MsgWritePaymentResponse{
				Uid:    msg.Uid,
				Type:   msg.Type,
				Raw:    make(map[string]string),
				Samuid: a.samuid,
				Seq:    msg.GetSeq(),
			})
		}
	case *MsgResponseCodeQR:
		if err := func() error {
			a.samuid = msg.SamUid
			data := msg.Value
			mess := struct {
				Type  string          `json:"t"`
				Value json.RawMessage `json:"v"`
			}{}
			data = bytes.TrimRight(data, "\x00")
			if err := json.Unmarshal(data, &mess); err != nil {
				return fmt.Errorf("QR error: %w", err)
			}
			logs.LogInfo.Printf("NewQR: %s", msg.Value)
			switch strings.ToUpper(mess.Type) {

			case "PMRT":
				rechQR := new(recharge.RechargeQR)
				if err := json.Unmarshal(mess.Value, rechQR); err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				logs.LogInfo.Printf("PMRT: %+v", rechQR)
				rechg := new(recharge.Recharge)
				date := time.Unix(rechQR.Date, 0)
				rechg.Date = date
				rechg.Exp = time.Duration(rechQR.Exp) * time.Second
				varSplit := strings.Split(rechQR.PID, "-")
				id := 0
				if len(varSplit) > 0 {
					id, _ = strconv.Atoi(varSplit[len(varSplit)-1])
				}
				rechg.PID = uint(id)
				rechg.TID = rechQR.TID
				rechg.Value = rechQR.Value

				logs.LogInfo.Printf("PMRT: %+v", rechg)

				if ctx.Parent() != nil {
					ctx.Send(ctx.Parent(), rechg)
				}
			case "DCIT":
				sendMsg := new(parameters.ConfigParameters)
				if err := json.Unmarshal(mess.Value, sendMsg); err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				logs.LogInfo.Printf("DCIT: %+v", sendMsg)
				if ctx.Parent() != nil {
					ctx.Send(ctx.Parent(), sendMsg)
				}
			case EQPM:
				raw, data, err := paym(mess.Value)
				if err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				uid := uint64(0)
				if v, ok := data["pid"]; ok {
					switch {
					case v.GetIntValue() > 0:
						uid = uint64(v.GetIntValue())
					case v.GetInt64Value() > int64(0):
						uid = uint64(v.GetInt64Value())
					case len(v.GetStringValue()) > 0:
						vv, _ := strconv.Atoi(v.GetStringValue())
						uid = uint64(vv)
					}
				} else {
					uid = uint64(time.Now().UnixNano() / 1_000_0000)
				}
				if ctx.Parent() != nil {
					ctx.Request(ctx.Parent(), &messages.MsgPayment{
						Type: EQPM,
						Data: data,
						Raw:  raw,
						Uid:  uid,
					})
				}
			case AQPM:
				raw, data, err := paym(mess.Value)
				if err != nil {
					return fmt.Errorf("QR error: %w", err)
				}
				uid := uint64(0)
				if v, ok := data["pid"]; ok {
					switch {
					case v.GetIntValue() > 0:
						uid = uint64(v.GetIntValue())
					case v.GetInt64Value() > int64(0):
						uid = uint64(v.GetInt64Value())
					case len(v.GetStringValue()) > 0:
						vv, _ := strconv.Atoi(v.GetStringValue())
						uid = uint64(vv)
					}
				} else {
					uid = uint64(time.Now().UnixNano() / 1_000_0000)
				}
				if ctx.Parent() != nil {
					ctx.Request(ctx.Parent(), &messages.MsgPayment{
						Type: AQPM,
						Data: data,
						Raw:  raw,
						Uid:  uid,
					})
				}
			}
			return nil
		}(); err != nil {
			logs.LogError.Printf(err.Error())
		}
	}
}
