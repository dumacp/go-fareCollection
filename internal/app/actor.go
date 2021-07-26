package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/dumacp/go-fareCollection/internal/buzzer"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/picto"
	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-fareCollection/pkg/payment/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment/mplus"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

type Actor struct {
	// lastTag        uint64
	// lastTimeDetect time.Time
	// errorWriteTag  uint64
	// actualTag      uint64
	pidGraph   *actor.PID
	pidBuzzer  *actor.PID
	pidPicto   *actor.PID
	pidReader  *actor.PID
	pidQR      *actor.PID
	inputs     int
	fmachine   *fsm.FSM
	lastTime   time.Time
	ctx        actor.Context
	mcard      payment.Payment
	updates    map[string]interface{}
	chNewRand  chan int
	lastRand   int
	actualRand int
}

func NewActor() actor.Actor {
	a := &Actor{}
	a.newFSM(nil)
	go a.RunFSM()
	a.chNewRand = make(chan int)
	go a.TickQR(a.chNewRand)
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		if err := func() error {
			propsGrpah := actor.PropsFromProducer(graph.NewActor)
			pidGrpah, err := ctx.SpawnNamed(propsGrpah, "graph-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidGraph = pidGrpah
			propsBuzzer := actor.PropsFromProducer(buzzer.NewActor)
			pidBuzzer, err := ctx.SpawnNamed(propsBuzzer, "buzzer-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidBuzzer = pidBuzzer
			propsPicto := actor.PropsFromProducer(picto.NewActor)
			pidPicto, err := ctx.SpawnNamed(propsPicto, "picto-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidPicto = pidPicto
			propsQR := actor.PropsFromProducer(qr.NewActor)
			pidQR, err := ctx.SpawnNamed(propsQR, "qr-actor")
			if err != nil {
				time.Sleep(3 * time.Second)
				return err
			}
			a.pidQR = pidQR
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			time.Sleep(3 * time.Second)
			panic(err)
		}
		a.fmachine.Event(eStarted)
		ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
	// case *MsgTagDetected:
	// 	if err := func() error {
	// 		if a.actualTag == a.lastTag {
	// 			if a.lastTimeDetect.Before(time.Now().Add(-5 * time.Second)) {
	// 				return nil
	// 			}
	// 		}
	// 		//read card
	// 		ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
	// 		a.lastTag = msg.UID
	// 		return nil
	// 	}(); err != nil {
	// 		logs.LogError.Println(err)
	// 	}
	case *messages.MsgPayment:

		jsonprint, err := json.MarshalIndent(msg.Data, "", "  ")
		if err != nil {
			logs.LogError.Println(err)
		}
		logs.LogBuild.Printf("tag read: %s", jsonprint)
		var paym payment.Payment
		if err := func() error {
			// if a.actualTag == a.errorWriteTag {
			// 	//Commit Tag
			// 	return nil
			// }
			if a.pidReader == nil {
				a.pidReader = ctx.Sender()
			}
			mcard := make(map[string]interface{})
			for k, v := range msg.Data {
				switch value := v.Data.(type) {
				case *messages.Value_Int64Value:
					mcard[k] = value.Int64Value
				case *messages.Value_Uint64Value:
					mcard[k] = value.Uint64Value
				case *messages.Value_IntValue:
					mcard[k] = int(value.IntValue)
				case *messages.Value_UintValue:
					mcard[k] = uint(value.UintValue)
				case *messages.Value_StringValue:
					mcard[k] = value.StringValue
				case *messages.Value_BytesValue:
					mcard[k] = value.BytesValue
				}
			}
			// v, err := payment.ValidationTag(a.mcard, 1028, 1290)
			paym = mplus.ParseToPayment(msg.Uid, mcard)
			if err := paym.AddBalance(-1200, 0000, 3033, 4044); err != nil {
				logs.LogBuild.Println(err)
				if errors.Is(err, payment.ErrorBalance) {
					//Send Msg Error Balance

					if balanceErr, ok := err.(*payment.ErrorBalanceValue); ok {
						a.fmachine.Event(eError, []string{`Saldo insuficiente`, fmt.Sprintf("%.02f", balanceErr.Balance)})
						// ctx.Send(a.pidGraph, &graph.MsgBalanceError{Value: fmt.Sprintf("%.02f", balanceErr.Balance)})
					} else {
						a.fmachine.Event(eError, []string{`Saldo insuficiente`, ""})
						// ctx.Send(a.pidGraph, &graph.MsgBalanceError{Value: ""})
					}
					// a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
					// a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
				}
				// time.Sleep(3 * time.Second)
				return err
			}
			a.updates = paym.Updates()
			update := make(map[string]*messages.Value)
			for k, val := range paym.Updates() {
				switch value := val.(type) {
				case int:
					update[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
				case uint:
					update[k] = &messages.Value{Data: &messages.Value_UintValue{UintValue: uint32(value)}}
				case int64:
					update[k] = &messages.Value{Data: &messages.Value_Int64Value{Int64Value: int64(value)}}
				case uint64:
					update[k] = &messages.Value{Data: &messages.Value_Uint64Value{Uint64Value: uint64(value)}}
				case string:
					update[k] = &messages.Value{Data: &messages.Value_StringValue{StringValue: value}}
				case []byte:
					update[k] = &messages.Value{Data: &messages.Value_BytesValue{BytesValue: value}}
				}
			}
			ctx.Request(ctx.Sender(), &messages.MsgWritePayment{Uid: msg.Uid, Updates: update})
			return nil
		}(); err != nil {
			a.fmachine.Event(eError, 1200)
			logs.LogError.Println(err)
		}
	case *messages.MsgWritePaymentResponse:
		// ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: fmt.Sprintf("$%.02f", float32(a.mcard["newSaldo"].(int32)))})
		a.fmachine.Event(eCardValidated, a.mcard.Balance())
		go func() {
			tID := a.mcard.Consecutive()
			// if !ok {
			// 	logs.LogError.Println("seq is not INT")
			// }
			id := a.mcard.ID()
			// if !ok {
			// 	logs.LogError.Println("\"name\" is not STRING")
			// }
			card := make(map[string]interface{})
			for k, v := range a.mcard.Data() {
				card[k] = v
			}
			for k, v := range a.mcard.Updates() {
				card[k] = v
			}
			// go func() {
			SendUsoTAG(fmt.Sprintf("%d", id), int(tID+1), card, a.mcard.Data(), []float64{0, 0}, time.Now())
			// if err != nil {
			// 	logs.LogError.Printf("POST error: %s", err)
			// 	return
			// }
			// logs.LogInfo.Printf("response platform: %s", response)
			// }()
		}()
	// case *MsgTagWriteError:
	// 	if err := func() error {
	// 		//Send Msg Write error
	// 		defer ctx.Send(nil, nil)
	// 		return nil
	// 	}(); err != nil {
	// 		logs.LogError.Println(err)
	// 	}
	case *qr.MsgNewCodeQR:
		logs.LogBuild.Printf("NewQR: %s", msg.Value)
		v, err := DecodeQR(msg.Value)
		if err != nil {
			logs.LogError.Println(err)
		}
		logs.LogBuild.Printf("NewQR: %s, [% X]", v, v)

		//TODO: fix this bug
		newv := make([]byte, 0)
		for i := range v {
			if v[i] > 0x20 {
				newv = append(newv, v[i])
			}
		}

		res := new(QrCode)
		if err := json.Unmarshal(newv, res); err != nil {
			logs.LogError.Printf("QR error: %s", err)
			break
		}
		pin, err := strconv.Atoi(res.Pin)
		if err != nil {
			logs.LogError.Printf("QR error: %s", err)
			break
		}
		if pin != a.lastRand && pin != a.actualRand {
			a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
			//TODO: cahngeeee!!!
			go func() {
				time.Sleep(2 * time.Second)
				a.ctx.Send(a.pidPicto, &picto.MsgPictoOFF{})
			}()
			logs.LogError.Printf("QR error: PIN is invalid")
			break
		}
		a.fmachine.Event(eQRValidated, fmt.Sprintf("%d", res.TransactionID))
		// ctx.Send(a.pidGraph, &graph.MsgValidationQR{Value: fmt.Sprintf("%d", res.TransactionID)})

		select {
		case a.chNewRand <- 1:
		case <-time.After(100 * time.Millisecond):
		}
		a.lastRand = a.actualRand
		a.actualRand = -1

		// go func() {
		SendUsoQR(int(res.TransactionID), []float64{0, 0}, time.Now())
		// 	if err != nil {
		// 		logs.LogError.Printf("POST error: %s", err)
		// 		return
		// 	}
		// 	logs.LogInfo.Printf("response platform: %s", response)
		// }()

	case *MsgNewRand:
		a.lastRand = a.actualRand
		a.actualRand = msg.Value
		v := fmt.Sprintf(urlQr, msg.Value)
		ctx.Send(a.pidGraph, &graph.MsgQrValue{Value: v})
	}
}
