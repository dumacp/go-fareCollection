package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"

	"github.com/dumacp/go-fareCollection/internal/business"
	"github.com/dumacp/go-fareCollection/internal/buzzer"
	"github.com/dumacp/go-fareCollection/internal/gps"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/picto"
	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-logs/pkg/logs"
	semessages "github.com/dumacp/go-sesam/pkg/messages"
	"github.com/looplab/fsm"
)

type Actor struct {
	deviceID    string
	version     string
	deviceIDnum int
	// lastTag        uint64
	// lastTimeDetect time.Time
	// errorWriteTag  uint64
	// actualTag      uint64
	pidGraph  *actor.PID
	pidBuzzer *actor.PID
	pidPicto  *actor.PID
	pidQR     *actor.PID
	pidFare   *actor.PID
	pidList   *actor.PID
	pidGps    *actor.PID
	pidUso    *actor.PID
	pidParams *actor.PID
	pidSam    *actor.PID
	fmachine  *fsm.FSM
	lastTime  time.Time
	ctx       actor.Context
	paym      map[uint64]payment.Payment
	recharge  *recharge.Recharge
	// rawcard   map[string]string
	// updates        map[string]interface{}
	// chNewRand  chan int
	oldRand int
	newRand int
	inputs  int
	outputs int
	seq     int
	// itineraryMap   itinerary.ItineraryMap
	params          *parameters.Parameters
	listRestrictive map[string]string
	timeout         int
	lastWriteError  uint64
	quit            chan int
}

func NewActor(id string, version string) actor.Actor {
	a := &Actor{}
	a.deviceID = id
	a.version = version
	varSplit := strings.Split(id, "-")
	if len(varSplit) > 0 {
		a.deviceIDnum, _ = strconv.Atoi(varSplit[len(varSplit)-1])
	}
	a.newFSM(nil)
	go a.RunFSM()
	return a
}

// var count = 0

func (a *Actor) Receive(ctx actor.Context) {
	a.ctx = ctx
	logs.LogBuild.Printf("Message arrived in appActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Stopping:
		ctx.Send(ctx.Self(), &MsgWriteAppParams{})
		ctx.Send(ctx.Self(), &MsgWriteErrorVerify{})
	case *actor.Started:
		if err := func() error {
			// a.params = new(parameters.Parameters)
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
		// ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)
	case *MsgWriteAppParams:
		if a.pidParams != nil {
			ctx.Send(a.pidParams, &parameters.AppParameters{
				Seq:     uint(a.seq),
				Inputs:  a.inputs,
				Outputs: a.outputs,
			})
		}
	case *parameters.ConfigParameters:
		if a.pidParams != nil {
			ctx.Send(a.pidParams, msg)
		}
		a.fmachine.Event(eOk, []string{"configuración", "datos actualizados"})
	case *parameters.MsgParameters:
		if ctx.Sender() != nil {
			a.pidParams = ctx.Sender()
		}
		a.params = msg.Data
		a.timeout = a.params.Timeout
		if a.params.Seq >= uint(a.seq) {
			a.seq = int(a.params.Seq)
		}
		if a.params.Inputs >= a.inputs {
			if a.pidGraph != nil {
				a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.params.Inputs})
			}
			a.inputs = int(a.params.Inputs)
		}
		if a.params.Outputs >= a.outputs {
			a.outputs = int(a.params.Outputs)
		}
		if a.params.DevSerial > 0 {
			if a.pidGraph != nil {
				a.ctx.Send(a.pidGraph, &graph.MsgRef{
					Device:  fmt.Sprintf("%s-%d", a.deviceID, a.deviceIDnum),
					Version: a.version,
					Ruta:    fmt.Sprintf("%d", a.params.PaymentItinerary),
				})
			}
			a.deviceIDnum = a.params.DevSerial
		}
		if a.pidList != nil {
			for _, list := range a.params.RestrictiveList {
				ctx.Request(a.pidList, &lists.MsgWatchList{ID: list})
			}
		}
		if a.pidGraph != nil {
			a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.deviceIDnum})
			a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.inputs})
		}
	case *lists.MsgWatchListResponse:
		if a.listRestrictive == nil {
			a.listRestrictive = make(map[string]string)
		}
		a.listRestrictive[msg.ID] = msg.PaymentMediumType
	case *messages.RegisterFareActor:
		a.pidFare = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterListActor:
		a.pidList = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterGPSActor:
		a.pidGps = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterUSOActor:
		a.pidUso = actor.NewPID(msg.Addr, msg.Id)
	case *messages.RegisterSAMActor:
		a.pidSam = actor.NewPID(msg.Addr, msg.Id)
	case *MsgWriteErrorVerify:
		if a.lastWriteError <= 0 {
			break
		}
		if len(a.paym) > 0 && a.paym[a.lastWriteError] != nil {
			paym := a.paym[a.lastWriteError]
			logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
			lasth := paym.Historical()

			tt := int64(0)
			if len(lasth) > 0 {
				tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
			}
			uso := &usostransporte.UsoTransporte{
				ID:                    paym.ID(),
				DeviceID:              a.deviceID,
				PaymentMediumTypeCode: paym.Type(),
				PaymentMediumId:       fmt.Sprintf("%d", paym.PID()),
				MediumID:              fmt.Sprintf("%d", paym.MID()),
				FareCode:              int(paym.FareID()),
				RawDataPrev:           paym.RawDataBefore(),
				RawDataAfter:          paym.RawDataAfter(),
				TransactionType:       "TRANSPORT_FARE_COLLECTION",
				TransactionTime:       tt,
				Error: &usostransporte.Error{
					Name: "write error",
					Desc: paym.Error(),
					Code: 1,
				},
				Coord: paym.Coord(),
			}
			ctx.Send(a.pidUso, uso)
		}
	case *usostransporte.UsoTransporte:
		logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
		if len(msg.Coord) <= 0 {
			coord := ""
			if a.pidGps != nil {
				res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
				if err != nil {
					logs.LogWarn.Printf("get fare err: %s", err)
				} else {
					switch v := res.(type) {
					case *gps.MsgGPS:
						if v.Data != nil {
							coord = string(v.Data)
							logs.LogBuild.Printf("coord: %s", coord)
						}
					}
				}
			}
			msg.Coord = coord
		}
		ctx.Send(a.pidUso, &usostransporte.MsgUso{
			Data: msg,
		})
	case *recharge.Recharge:
		logs.LogInfo.Printf("recharge: %+v", msg)
		if msg.Exp.Before(time.Now()) {
			err := NewErrorScreen("error en la recarga", "recarga expirada")
			a.fmachine.Event(eError, err)
			logs.LogError.Println(err)
			a.recharge = nil
			break
		}
		a.fmachine.Event(eOk, []string{"ticket de recarga", fmt.Sprintf("(%d) ubique la tarjeta", msg.Value)})
		a.recharge = msg
		a.recharge.Seq = a.seq + 1
		a.recharge.DeviceID = a.deviceIDnum
	case *messages.MsgPayment:
		/**/

		updates, err := func() (map[string]interface{}, error) {
			logs.LogInfo.Printf("detect uid: %d", msg.Uid)
			paym, err := business.ParsePayment(msg)
			logs.LogInfo.Printf("parse uid: %d", msg.Uid)
			if err != nil {
				return nil, err
			}
			ok, err := business.VerifyListRestrictive(ctx, a.pidList, paym.Type(), int64(paym.PID()), a.listRestrictive)
			if err != nil {
				logs.LogWarn.Println(err)
			}

			if ok {
				return nil, NewErrorScreen("tarjeta bloqueada", "tarjeta en listas restrictivas")
			}

			if a.paym == nil {
				a.paym = make(map[uint64]payment.Payment)
			}

			switch paym.Type() {
			case "MIFARE_PLUS_EV2_4K":
				logs.LogInfo.Printf("rewrite try, uid: %d, last: %d", msg.Uid, a.lastWriteError)
				lastPaym := a.paym[msg.GetUid()]
				if business.Rewrite(paym, lastPaym, a.lastWriteError) {
					logs.LogInfo.Printf("rewrite try, uid: %d", msg.Uid)
					logs.LogInfo.Printf("payment last (%d) updates: %+v", lastPaym.PID(), lastPaym.Updates())
					a.paym[msg.GetUid()] = lastPaym
					return lastPaym.Updates(), nil
				}
				logs.LogInfo.Printf("payment (%d) updates: %+v", paym.PID(), paym.Updates())
				a.paym[msg.GetUid()] = paym
				if lastPaym != nil && paym.MID() != lastPaym.MID() {
					lasth := lastPaym.Historical()

					tt := int64(0)
					if len(lasth) > 0 {
						tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
					}
					uso := &usostransporte.UsoTransporte{
						ID:                    lastPaym.ID(),
						DeviceID:              a.deviceID,
						PaymentMediumTypeCode: msg.Type,
						PaymentMediumId:       fmt.Sprintf("%d", lastPaym.PID()),
						MediumID:              fmt.Sprintf("%d", lastPaym.MID()),
						FareCode:              int(lastPaym.FareID()),
						RawDataPrev:           lastPaym.RawDataBefore(),
						RawDataAfter:          lastPaym.RawDataAfter(),
						TransactionType:       "TRANSPORT_FARE_COLLECTION",
						TransactionTime:       tt,
						Error: &usostransporte.Error{
							Name: "write error",
							Desc: lastPaym.Error(),
							Code: 1,
						},
					}
					ctx.Send(ctx.Self(), uso)
				}
				if a.recharge != nil {
					defer func() {
						a.recharge = nil
					}()
					if _, err := business.RechargeQR(paym, a.recharge); err != nil {
						return nil, NewErrorScreen("error recarga", err.Error())
					}
				}
				updates, err := business.CalcUpdatesWithFare(ctx, a.pidFare, a.deviceIDnum, paym, a.params)
				if err != nil {
					if errors.Is(err, payment.ErrorBalance) {
						return nil, NewErrorScreen("saldo insuficiente", err.Error())
					} else {
						return nil, err
					}
				}
				return updates, nil
			case "ENDUSER_QR":
				logs.LogInfo.Printf("payment (%d) updates: %+v", paym.PID(), paym.Updates())
				a.paym[msg.GetUid()] = paym
				return business.CalcUpdatesQR(paym, &a.oldRand, &a.newRand)
			default:
				logs.LogInfo.Printf("payment (%d) updates: %+v", paym.PID(), paym.Updates())
				a.paym[msg.GetUid()] = paym
				return nil, fmt.Errorf("payment type not found: %s, %s", msg.Type, paym.Type())
			}
		}()
		if err != nil {
			if a.paym != nil {
				delete(a.paym, msg.GetUid())
			}
			a.fmachine.Event(eError, err)
			logs.LogError.Println(err)
			break
		}

		updateValues := business.ParseUpdatesToValuePayment(updates)

		if ctx.Sender() != nil {
			sendMsg := &messages.MsgWritePayment{
				Uid:     msg.Uid,
				Type:    msg.GetType(),
				Updates: updateValues,
				Seq:     int32(a.seq + 1),
			}
			if a.lastWriteError == msg.GetUid() {
				sendMsg.AlreadyUpdate = true
			}
			ctx.Request(ctx.Sender(), sendMsg)
		}
		a.lastWriteError = 0
		/**/
	case *messages.MsgWritePaymentError:
		a.lastWriteError = msg.GetUid()
		if a.paym == nil {
			a.paym = make(map[uint64]payment.Payment)
		}
		if a.paym[msg.GetUid()] == nil {
			break
		}
		lastPaym := a.paym[msg.GetUid()]
		raw := msg.GetRaw()
		raw["mv"] = fmt.Sprintf("%d", lastPaym.VersionLayout())
		switch msg.GetType() {
		case "MIFARE_PLUS_EV2_4K":
			raw["mv"] = "3"
		}
		lastPaym.SetRawDataAfter(raw)
		lastPaym.SetError(msg.Error)
		coord := ""
		if a.pidGps != nil {
			res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
			if err != nil {
				logs.LogWarn.Printf("get fare err: %s", err)
			} else {
				switch v := res.(type) {
				case *gps.MsgGPS:
					if v.Data != nil {
						coord = string(v.Data)
						logs.LogBuild.Printf("coord: %s", coord)
					}
				}
			}
		}
		lastPaym.SetCoord(coord)
		a.fmachine.Event(eError, NewErrorScreen("error de escritura", "vuelva a ubicar la tarjeta"))
		logs.LogError.Printf("error de escritura, uid: %d, err: %s", msg.Uid, msg.Error)
	case *messages.MsgWritePaymentResponse:
		if a.paym == nil {
			a.paym = make(map[uint64]payment.Payment)
		}
		if a.paym[msg.GetUid()] == nil {
			break
		}
		lastPaym := a.paym[msg.GetUid()]
		a.inputs++
		a.seq++

		logs.LogInfo.Printf("payment write map: %+v, updates: %+v", lastPaym.Data(), lastPaym.Updates())
		// ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: fmt.Sprintf("$%.02f", float32(a.mcard["newSaldo"].(int32)))})
		a.fmachine.Event(eCardValidated, lastPaym.Balance())

		cardid := lastPaym.PID()
		uid := msg.Uid

		lasth := lastPaym.Historical()
		tt := int64(0)
		if len(lasth) > 0 {
			tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
		}

		raw := msg.GetRaw()
		raw["mv"] = fmt.Sprintf("%d", lastPaym.VersionLayout())
		switch msg.GetType() {
		case "MIFARE_PLUS_EV2_4K":
			raw["mv"] = "3"
		}
		lastPaym.SetRawDataAfter(raw)

		uso := &usostransporte.UsoTransporte{
			ID:                     lastPaym.ID(),
			DeviceID:               a.deviceID,
			SamUuid:                msg.Samuid,
			PaymentMediumTypeCode:  msg.Type,
			PaymentMediumId:        fmt.Sprintf("%d", cardid),
			MediumID:               fmt.Sprintf("%d", uid),
			FareCode:               int(lastPaym.FareID()),
			TerminalTransactionSeq: int(msg.Seq),
			RawDataPrev:            lastPaym.RawDataBefore(),
			RawDataAfter:           lastPaym.RawDataAfter(),
			TransactionType:        "TRANSPORT_FARE_COLLECTION",
			TransactionTime:        tt,
			Error: &usostransporte.Error{
				Name: "",
				Desc: "",
				Code: 0,
			},
		}
		switch msg.Type {
		case "MIFARE_PLUS_EV2_4K":
			hr := lastPaym.Recharged()
			// for i, v := range hr {
			// 	logs.LogInfo.Printf("last hist %d: %+v", i, v)
			// }
			// logs.LogInfo.Printf("msg seq: %d", msg.Seq)

			if len(hr) > 0 && hr[len(hr)-1].ConsecutiveID() == uint(msg.GetSeq()) &&
				hr[len(hr)-1].DeviceID() == uint(a.deviceIDnum) {
				logs.LogInfo.Printf("payment recharged: %+v", hr[len(hr)-1])
				uso.TransactionType = "TFC_WITH_BALANCE_RECHARGE"
				// uso.RechargeTokenId = int(a.recharge.TID)
				// uso.RechargeValue = a.recharge.Value
				// logs.LogInfo.Printf("last hist: %+v", hr[len(hr)-1])
				if v := hr[len(hr)-1].RechargeProp("RechargeTokenId"); v != nil {
					if tid, ok := v.(int); ok {
						uso.RechargeTokenId = int(tid)
						uso.RechargeValue = hr[len(hr)-1].Value()
					}
				}
			}
		}
		ctx.Send(ctx.Self(), uso)
		a.lastWriteError = 0
		delete(a.paym, msg.GetUid())
	case *qr.MsgNewCodeQR:
		if err := func() error {
			divInput := msg.Value[0:4]
			revDiv := make([]byte, 0)
			for i := range divInput {
				revDiv = append(revDiv, divInput[len(divInput)-1-i])
			}
			iv := make([]byte, 0)
			iv = append(iv, divInput...)
			iv = append(iv, revDiv...)
			iv = append(iv, divInput...)
			iv = append(iv, revDiv...)
			mdec := &semessages.MsgDecryptRequest{
				Data:     msg.Value[4:],
				DevInput: divInput[0:4],
				IV:       iv,
				//TODO: How to get ?
				KeySlot: 0x33,
			}
			logs.LogBuild.Printf("QR crypt: [% X], len: %d; [% X], len: %d",
				mdec.Data, len(mdec.Data), mdec.DevInput, len(mdec.DevInput))

			var data []byte

			res, err := ctx.RequestFuture(a.pidSam, mdec, time.Millisecond*600).Result()
			if err != nil {
				return fmt.Errorf("get decrypt sam err: %w", err)
			}
			switch v := res.(type) {
			case *semessages.MsgDecryptResponse:
				if v.Plain == nil {
					return errors.New("QR decrypt data is empty")
				}
				data = v.Plain
				logs.LogBuild.Printf("QR decrypt: %s, [%X]", data, data)
			}

			if ctx.Sender() != nil {
				ctx.Respond(&qr.MsgResponseCodeQR{
					Value: data,
				})
			}
			return nil
		}(); err != nil {
			logs.LogError.Printf("QR error: %s", err)
		}
	case *qr.MsgNewRand:
		a.oldRand = a.newRand
		a.newRand = msg.Value
		var v string
		if a.params == nil {
			v = "https://fleet.nebulae.com.co/siv"
		} else {
			v = fmt.Sprintf(urlQr, a.params.PaymentItinerary, a.deviceID, msg.Value)
		}
		ctx.Send(a.pidGraph, &graph.MsgQrValue{Value: v})
	}
}
