package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/dumacp/go-fareCollection/internal/business"
	"github.com/dumacp/go-fareCollection/internal/gps"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/lock"
	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-logs/pkg/logs"
)

func (a *Actor) RunState(ctx actor.Context) {
	a.ctx = ctx
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		logs.LogInfo.Printf("started \"%s\", \"RunState\", %v", ctx.Self().GetId(), ctx.Self())
	case *parameters.MsgParameters:
		logstrans.LogInfo.Printf("params: %+v", msg.Data)
		if ctx.Sender() != nil {
			a.pidParams = ctx.Sender()
		}

		a.params = msg.Data
		a.timeout = a.params.Timeout
		if a.params.Seq >= uint(a.seq) {
			a.seq = int(a.params.Seq)
		}
		if a.params.Inputs >= a.inputs {
			a.inputs = int(a.params.Inputs)
		}
		if a.params.Outputs >= a.outputs {
			a.outputs = int(a.params.Outputs)
		}
		if a.params.DevSerial > 0 {
			a.deviceIDnum = a.params.DevSerial
		}
		if a.pidList != nil {
			for _, list := range a.params.RestrictiveList {
				ctx.Request(a.pidList, &lists.MsgWatchList{ID: list})
			}
		}
		if a.pidGraph != nil {
			a.ctx.Send(a.pidGraph, &graph.MsgRef{
				Device:  a.deviceID,
				Version: a.version,
				Ruta:    fmt.Sprintf("Ruta: %d", a.params.PaymentItinerary),
			})
			a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.inputs})

		}
	case *MsgWriteErrorVerify:
		a.writeerrorverify()
	case *usostransporte.UsoTransporte:
		// logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
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
	case *lock.Lock:
		// logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
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
		ctx.Send(a.pidUso, &usostransporte.MsgLock{
			Data: msg,
		})
	case *recharge.Recharge:
		logs.LogInfo.Printf("recharge: %+v", msg)
		if msg.Date.Add(msg.Exp).Before(time.Now()) {
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
		// if a.disableApp {
		// 	a.fmachine.Event(eError, NewErrorScreen("error de sistema", "vuelva a ubicar la tarjeta"))
		// }
		var paym payment.Payment
		if a.paym == nil {
			a.paym = make(map[uint64]payment.Payment)
		}
		updates, err := func() (map[string]interface{}, error) {
			logstrans.LogBuild.Printf("detect uid: %d", msg.Uid)
			var err error
			paym, err = business.ParsePayment(msg)
			if err != nil {
				return nil, err
			}
			logstrans.LogBuild.Printf("parse uid: %d", msg.Uid)
			if paym.Lock() {
				logstrans.LogWarn.Printf("tag locked, uid: %X, id: %d", paym.MID(), paym.PID())
				return nil, NewErrorScreen("tarjeta bloqueada", "dirígase a un PAC")
			}
			ok, codeList, err := business.VerifyListRestrictive(ctx, a.pidList, paym.Type(), int64(paym.PID()), a.listRestrictive)
			if err != nil {
				logstrans.LogWarn.Printf("restrictive list err: %s", err)
			}
			if ok {
				logstrans.LogWarn.Printf("tag in restrictive list: %s, %d", paym.Type(), paym.PID())
				paym.SetLock(lock.ON_RESTRICTIVE_LIST.String(), codeList, a.listRestrictive[codeList].Version)
				a.paym[msg.GetUid()] = paym
				return paym.Updates(), NewErrorScreen("tarjeta bloqueada", "tarjeta en listas restrictivas")
			}
			switch paym.Type() {
			case "MIFARE":
				// logstrans.LogInfo.Printf("rewrite try, uid: %d, last: %d", msg.Uid, a.lastWriteError)
				lastPaym := a.paym[msg.GetUid()]
				if business.Rewrite(paym, lastPaym, a.lastWriteError) {
					logstrans.LogInfo.Printf("rewrite try, uid: %d", msg.Uid)
					logstrans.LogBuild.Printf("payment last (%d) updates: %+v", lastPaym.PID(), lastPaym.Updates())
					a.paym[msg.GetUid()] = lastPaym
					return lastPaym.Updates(), nil
				}
				// if lastPaym != nil {
				// 	logstrans.LogInfo.Printf("rewrite try, uid: %d, last uid: %d", msg.Uid, lastPaym.MID())
				// }
				// logstrans.LogBuild.Printf("payment (%d) updates: %+v", paym.PID(), paym.Updates())

				if a.lastWriteError > 0 && paym.MID() != a.lastWriteError {
					lastPaym = a.paym[a.lastWriteError]
					if lastPaym != nil {
						defer func() {
							if len(a.paym) > 0 {
								delete(a.paym, a.lastWriteError)
							}
						}()
						logstrans.LogInfo.Printf("cancel rewrite, uid: %d", msg.Uid)
						lasth := lastPaym.Historical()

						tt := int64(0)
						iti := uint(0)
						if len(lasth) > 0 {
							tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
							iti = lasth[len(lasth)-1].ItineraryID()
						}
						uso := &usostransporte.UsoTransporte{
							ID:                    lastPaym.ID(),
							DeviceID:              a.deviceID,
							ItineraryID:           int(iti),
							PaymentMediumTypeCode: msg.Type,
							PaymentMediumId:       fmt.Sprintf("%d", lastPaym.PID()),
							MediumID:              fmt.Sprintf("%X", lastPaym.MID()),
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
					} else {
						logstrans.LogError.Printf("cancel rewrite without data, uid: %d", msg.Uid)
					}
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
				a.paym[msg.GetUid()] = paym
				return updates, nil
			case qr.EQPM:
				itiID := uint(0)
				if a.params != nil {
					itiID = uint(a.params.PaymentItinerary)
				}
				if _, err := business.CalcUpdatesQR(paym, itiID, &a.oldRand, &a.newRand); err != nil {
					return nil, NewErrorScreen("ticket no válido", err.Error())
				}
				a.paym[msg.GetUid()] = paym
				return nil, nil
			case qr.AQPM:
				itiID := uint(0)
				if a.params != nil {
					itiID = uint(a.params.PaymentItinerary)
				}
				if _, err := business.CalcUpdatesAnonQR(paym, itiID); err != nil {
					return nil, NewErrorScreen("ticket no válido", err.Error())
				}
				a.paym[msg.GetUid()] = paym
				return nil, nil
			default:
				a.paym[msg.GetUid()] = paym
				return nil, fmt.Errorf("payment type not found: %s, %s", msg.Type, paym.Type())
			}
		}()
		if err != nil {

			a.fmachine.Event(eError, err)
			if paym != nil {
				logstrans.LogError.Printf("payment err: uid: %X, id: %d, err: %s", paym.MID(), paym.PID(), err)
			}
			if updates == nil {
				if len(a.paym) > 0 {
					delete(a.paym, msg.GetUid())
				}
				if ctx.Sender() != nil {
					go func(ctxxx actor.Context, sender *actor.PID) {
						time.Sleep(1 * time.Second)
						sendMsg := &messages.MsgDetectPayment{}
						ctxxx.Request(sender, sendMsg)
					}(ctx, ctx.Sender())
				}
				break
			}
		}

		updateValues := business.ParseUpdatesToValuePayment(updates)

		if ctx.Sender() != nil {
			sendMsg := &messages.MsgWritePayment{
				Uid:     msg.GetUid(),
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
		case "MIFARE":
			raw["mv"] = "3"
		}
		lastPaym.SetRawDataAfter(raw)
		lastPaym.SetError(msg.Error)
		coord := ""
		if a.pidGps != nil {
			res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
			if err != nil {
				logstrans.LogWarn.Printf("get fare err: %s", err)
			} else {
				switch v := res.(type) {
				case *gps.MsgGPS:
					if v.Data != nil {
						coord = string(v.Data)
						logstrans.LogBuild.Printf("coord: %s", coord)
					}
				}
			}
		}
		lastPaym.SetCoord(coord)
		a.fmachine.Event(eError, NewErrorScreen("error de escritura", "vuelva a ubicar la tarjeta"))
		logstrans.LogWarn.Printf("write tag error: uid: %X, id: %d, err: %s",
			msg.Uid, lastPaym.PID(), msg.Error)
	case *messages.MsgWritePaymentResponse:
		func() {
			if a.paym == nil {
				a.paym = make(map[uint64]payment.Payment)
			}
			if a.paym[msg.GetUid()] == nil {
				return
			}
			defer func() {
				a.lastWriteError = 0
				if len(a.paym) > 0 {
					delete(a.paym, msg.GetUid())
				}
			}()

			lastPaym := a.paym[msg.GetUid()]
			cardid := lastPaym.PID()
			uid := msg.Uid
			tt := time.Now().UnixNano() / 1_000_000

			raw := msg.GetRaw()
			raw["mv"] = fmt.Sprintf("%d", lastPaym.VersionLayout())

			lastPaym.SetRawDataAfter(raw)

			if lastPaym.Lock() {
				//TODO: ?
				raw["mv"] = "3"
				lockk := &lock.Lock{
					ID:                       lastPaym.ID(),
					DeviceID:                 a.deviceID,
					SamUuid:                  msg.Samuid,
					PaymentMediumTypeCode:    msg.Type,
					PaymentMediumId:          fmt.Sprintf("%d", cardid),
					MediumID:                 fmt.Sprintf("%X", uid),
					RawDataPrev:              lastPaym.RawDataBefore(),
					RawDataAfter:             lastPaym.RawDataAfter(),
					Reason:                   lastPaym.LockReason(),
					PaymentMediumListVersion: lastPaym.LockListVersion(),
					PaymentMediumListId:      lastPaym.LockList(),
					TransactionTime:          tt,
					Error: &lock.Error{
						Name: "",
						Desc: "",
						Code: 0,
					},
				}
				ctx.Send(ctx.Self(), lockk)
				return
			}
			a.inputs++
			a.seq++

			// ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: fmt.Sprintf("$%.02f", float32(a.mcard["newSaldo"].(int32)))})
			// a.fmachine.Event(eCardValidated, lastPaym.Balance())

			lasth := lastPaym.Historical()
			// logs.LogBuild.Printf("lasth: %+v", lasth)

			iti := uint(0)
			if len(lasth) > 0 {
				// logs.LogInfo.Printf("lasth: %+v", lasth[len(lasth)-1])
				tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
				iti = lasth[len(lasth)-1].ItineraryID()
			}

			// raw := msg.GetRaw()
			// raw["mv"] = fmt.Sprintf("%d", lastPaym.VersionLayout())
			switch msg.GetType() {
			case "MIFARE":
				//TODO: ?
				raw["mv"] = "3"
				a.fmachine.Event(eCardValidated, lastPaym.Balance())
			default:
				a.fmachine.Event(eQRValidated, lastPaym.Balance())
			}

			uso := &usostransporte.UsoTransporte{
				ID:                     lastPaym.ID(),
				DeviceID:               a.deviceID,
				ItineraryID:            int(iti),
				SamUuid:                msg.Samuid,
				PaymentMediumTypeCode:  msg.Type,
				PaymentMediumId:        fmt.Sprintf("%d", cardid),
				MediumID:               fmt.Sprintf("%X", uid),
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
			case "MIFARE":
				hr := lastPaym.Recharged()
				// for i, v := range hr {
				// 	logs.LogInfo.Printf("last hist %d: %+v", i, v)
				// }
				// logs.LogInfo.Printf("msg seq: %d", msg.Seq)

				if len(hr) > 0 &&
					hr[len(hr)-1].DeviceID() == uint(a.deviceIDnum) &&
					hr[len(hr)-1].RechargeProp("Seq") != nil {
					seq, ok := hr[len(hr)-1].RechargeProp("Seq").(int)
					if ok && int32(seq) == msg.GetSeq() {
						logstrans.LogInfo.Printf("payment recharged: %+v", hr[len(hr)-1])
						uso.TransactionType = "TFC_WITH_BALANCE_RECHARGE"
						if v := hr[len(hr)-1].RechargeProp("RechargeTokenId"); v != nil {
							if tid, ok := v.(int); ok {
								uso.RechargeTokenId = int(tid)
								uso.RechargeValue = hr[len(hr)-1].Value()
							}
						}
					}
				}
			}
			ctx.Send(ctx.Self(), uso)
		}()

	case *qr.MsgNewRand:
		a.oldRand = a.newRand
		a.newRand = msg.Value
		var v string
		if a.params == nil || len(a.params.UrlQR) <= 0 {
			v = fmt.Sprintf("%s/siv", utils.Url)
		} else {
			v = fmt.Sprintf(qr.UrlQRformat, a.params.UrlQR, a.params.PaymentItinerary, a.deviceID, msg.Value)
		}
		if a.pidGraph != nil {
			ctx.Send(a.pidGraph, &graph.MsgQrValue{Value: v})
		}
	case *usostransporte.MsgErrorDB:
		a.fmachine.Event(eError, NewErrorScreen("error de sistema", "base de datos local error"))
		a.behavior.Become(a.AwaitState)
	case *messages.MsgSEError:
		a.fmachine.Event(eError, NewErrorScreen("error de sistema", "Security Element error"))
		logstrans.LogError.Printf("--- SE error, err: %s", msg.Error)
		a.behavior.Become(a.AwaitState)
		if a.pidUso != nil {
			ctx.Request(a.pidUso, &usostransporte.MsgVerifyDB{})
		}
	}
}
