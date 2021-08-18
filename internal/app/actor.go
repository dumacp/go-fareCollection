package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/google/uuid"

	"github.com/dumacp/go-fareCollection/internal/buzzer"
	"github.com/dumacp/go-fareCollection/internal/fare"
	"github.com/dumacp/go-fareCollection/internal/gps"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/internal/picto"
	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-fareCollection/pkg/payment/mplus"
	"github.com/dumacp/go-fareCollection/pkg/payment/token"
	"github.com/dumacp/go-logs/pkg/logs"
	semessages "github.com/dumacp/go-sesam/pkg/messages"
	"github.com/looplab/fsm"
)

type Actor struct {
	deviceID    string
	deviceIDnum int
	// lastTag        uint64
	// lastTimeDetect time.Time
	// errorWriteTag  uint64
	// actualTag      uint64
	pidGraph  *actor.PID
	pidBuzzer *actor.PID
	pidPicto  *actor.PID
	pidReader *actor.PID
	pidQR     *actor.PID
	pidFare   *actor.PID
	pidList   *actor.PID
	pidGps    *actor.PID
	pidUso    *actor.PID
	pidParams *actor.PID
	pidSam    *actor.PID
	inputs    int
	fmachine  *fsm.FSM
	lastTime  time.Time
	ctx       actor.Context
	mcard     payment.Payment
	rawcard   map[string]string
	// updates        map[string]interface{}
	// chNewRand  chan int
	oldRand int
	newRand int
	// itineraryMap   itinerary.ItineraryMap
	params          *parameters.Parameters
	listRestrictive map[string]string
	timeout         int
	lastWriteError  uint64
	quit            chan int
}

func NewActor(id string) actor.Actor {
	a := &Actor{}
	a.deviceID = id
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
		ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)
	case *parameters.ConfigParameters:
		if a.pidParams != nil {
			ctx.Send(a.pidParams, msg)
		}
	case *parameters.MsgParameters:
		if ctx.Sender() != nil {
			a.pidParams = ctx.Sender()
		}
		a.params = msg.Data
		a.timeout = a.params.Timeout
		if a.pidList != nil {
			for _, list := range a.params.RestrictiveList {
				ctx.Request(a.pidList, &lists.MsgWatchList{ID: list})
			}
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
		if a.lastWriteError > 0 && a.mcard != nil {
			logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
			id, err := uuid.NewUUID()
			if err != nil {
				logs.LogError.Printf("uuid err: %s", err)
				break
			}
			coord := ""
			if a.pidGps != nil {
				res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
				if err != nil {
					logs.LogWarn.Println("get fare err: %w", err)
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
			uso := &usostransporte.UsoTransporte{
				ID:                    id.String(),
				DeviceID:              a.deviceID,
				PaymentMediumTypeCode: a.mcard.Type(),
				PaymentMediumId:       fmt.Sprintf("%d", a.mcard.ID()),
				MediumID:              fmt.Sprintf("%d", a.mcard.UID()),
				FareCode:              int(a.mcard.FareID()),
				RawDataPrev:           a.mcard.RawDataBefore(),
				RawDataAfter:          a.mcard.RawDataAfter(),
				TransactionType:       "TRANSPORT_FARE_COLLECTION",
				Error: &usostransporte.Error{
					Name: "write error",
					Desc: a.mcard.Error(),
					Code: 1,
				},
				Coord: coord,
			}
			ctx.Send(a.pidUso, &usostransporte.MsgUso{
				Data: uso,
			})
		}
	case *messages.MsgPayment:
		/**/
		jsonprint, err := json.MarshalIndent(msg.Data, "", "  ")
		if err != nil {
			logs.LogError.Println(err)
		}
		logs.LogBuild.Printf("tag read: %s", jsonprint)

		/**/
		var paym payment.Payment
		update, err := func() (mappp map[string]interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					logs.LogError.Println("Recovered in *messages.MsgPayment app\", ", r)
					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = errors.New("unknown panic")
					}
					time.Sleep(3 * time.Second)
				}
			}()
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
			logs.LogBuild.Printf("tag map: %v", mcard)
			// v, err := payment.ValidationTag(a.mcard, 1028, 1290)

			switch msg.GetType() {
			case "MIFARE_PLUS_EV2_4K":
				paym = mplus.ParseToPayment(msg.Uid, msg.GetType(), mcard)
				logs.LogInfo.Printf("tag read uid: %d, lastErr: %d", paym.UID(), a.lastWriteError)
				if rewrite(paym, a.mcard, a.lastWriteError) {
					logs.LogInfo.Printf("rewrite try, uid: %d", msg.Uid)
					return a.mcard.Updates(), nil
				}
				if a.mcard != nil && paym.UID() != a.mcard.UID() {
					logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
					id, err := uuid.NewUUID()
					if err != nil {
						logs.LogError.Printf("uuid err: %s", err)
						break
					}
					coord := ""
					if a.pidGps != nil {
						res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
						if err != nil {
							logs.LogWarn.Println("get fare err: %w", err)
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
					uso := &usostransporte.UsoTransporte{
						ID:                    id.String(),
						DeviceID:              a.deviceID,
						PaymentMediumTypeCode: msg.Type,
						PaymentMediumId:       fmt.Sprintf("%d", a.mcard.ID()),
						MediumID:              fmt.Sprintf("%d", a.mcard.UID()),
						FareCode:              int(a.mcard.FareID()),
						RawDataPrev:           a.mcard.RawDataBefore(),
						RawDataAfter:          a.mcard.RawDataAfter(),
						TransactionType:       "TRANSPORT_FARE_COLLECTION",
						Error: &usostransporte.Error{
							Name: "write error",
							Desc: a.mcard.Error(),
							Code: 1,
						},
						Coord: coord,
					}
					ctx.Send(a.pidUso, &usostransporte.MsgUso{
						Data: uso,
					})
				}

			case "ENDUSER_QR":
				paym = token.ParseToPayment(msg.Uid, mcard)
			}

			raw := msg.GetRaw()
			raw["mv"] = fmt.Sprintf("%d", paym.VersionLayout())
			paym.SetRawDataBefore(raw)
			a.mcard = paym
			a.rawcard = msg.Raw
			logs.LogBuild.Printf("tag map parse: %+v", a.mcard.Data())
			//TODO: List
			if a.params != nil {
				for list, code := range a.listRestrictive {
					if code != msg.GetType() {
						continue
					}
					resList, err := ctx.RequestFuture(a.pidList, &lists.MsgVerifyInList{
						ListID: list,
						ID:     []int64{int64(a.mcard.ID())},
					}, 60*time.Millisecond).Result()
					if err != nil {
						logs.LogWarn.Printf("get restrictive list err: %s", err)
						break
					}
					switch v := resList.(type) {
					case *lists.MsgVerifyInListResponse:
						if len(v.ID) > 0 {
							logs.LogWarn.Printf("WARN!!! id in LIST: %d", v.ID)
							return nil, NewErrorScreen("tarjeta bloqueada", "tarjeta en listas restrictivas")
							// return NewErrorScreen("saldo insuficiente", fmt.Sprintf("%.02f", 1850.00))
						}
					}
				}
			}

			switch a.mcard.Type() {
			case "MIFARE_PLUS_EV2_4K":
				lastFares := make(map[int64]int)
				hs := a.mcard.Historical()
				for _, v := range hs {
					timestamp := v.TimeTransaction().Unix()
					fareid := v.FareID()
					lastFares[timestamp] = int(fareid)
				}
				getFare := &fare.MsgGetFare{
					LastFarePolicies: lastFares,
					ProfileID:        int(a.mcard.ProfileID()),
					// ItineraryID:      157,
					// ModeID:           1,
					// RouteID:         77,
					FromItineraryID: int(hs[len(hs)-1].ItineraryID()),
				}
				if a.params == nil {
					return nil, errors.New("params is nil")
				}

				getFare.ModeID = int(a.params.PaymentMode)
				//TODO: get ids by QR
				getFare.RouteID = int(a.params.PaymentRoute)
				getFare.ItineraryID = int(a.params.PaymentItinerary)

				//TODO: farePID?
				if a.pidFare == nil {
					return nil, errors.New("pidFare not found")
				}
				resFare, err := ctx.RequestFuture(a.pidFare, getFare, 30*time.Millisecond).Result()
				if err != nil {
					return nil, fmt.Errorf("get fare err: %w", err)
				}
				// cost := 0
				// fareID := 0
				var fareData *fare.MsgFare
				switch res := resFare.(type) {
				case *fare.MsgFare:

					// cost = res.Fare
					// fareID = res.FarePolicyID
					fareData = res
					//TODO: how get deviceID?
					fareData.DeviceID = a.deviceIDnum
					fareData.ItineraryID = a.params.PaymentItinerary
				case *fare.MsgError:
					return nil, errors.New(res.Err)
				default:
					return nil, errors.New("fareID not found")
				}

				logs.LogBuild.Printf("fare calc: %+v", fareData)

				if _, err := a.mcard.ApplyFare(fareData); err != nil {
					logs.LogBuild.Println(err)
					if errors.Is(err, payment.ErrorBalance) {
						//Send Msg Error Balance

						if balanceErr, ok := err.(*payment.ErrorBalanceValue); ok {
							return nil, NewErrorScreen("saldo insuficiente", fmt.Sprintf("%.02f", balanceErr.Balance))
						} else {
							return nil, NewErrorScreen("saldo insuficiente")
						}
						// a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
						// a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
					}
					// time.Sleep(3 * time.Second)
					return nil, err
				}
				// a.updates = paym.Updates()
				logs.LogBuild.Printf("tag updates: %+v", a.mcard.Updates())

				return a.mcard.Updates(), nil

			case "ENDUSER_QR":
				if pin, err := a.mcard.ApplyFare([]int{a.newRand, a.oldRand}); err != nil {
					a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
					a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
					go func() {
						time.Sleep(2 * time.Second)
						a.ctx.Send(a.pidPicto, &picto.MsgPictoOFF{})
					}()
					return nil, err
				} else {
					for _, v := range []*int{&a.newRand, &a.oldRand} {
						if *v == pin {
							*v = -1
						}
					}
				}

				return nil, nil

			}
			return nil, fmt.Errorf("payment type not found: %s, %s", msg.Type, a.mcard.Type())
		}()
		if err != nil {
			a.fmachine.Event(eError, err)
			logs.LogError.Println(err)
			break
		}
		updateValue := make(map[string]*messages.Value)
		for k, val := range update {
			switch value := val.(type) {
			case int:
				updateValue[k] = &messages.Value{Data: &messages.Value_IntValue{IntValue: int32(value)}}
			case uint:
				updateValue[k] = &messages.Value{Data: &messages.Value_UintValue{UintValue: uint32(value)}}
			case int64:
				updateValue[k] = &messages.Value{Data: &messages.Value_Int64Value{Int64Value: int64(value)}}
			case uint64:
				updateValue[k] = &messages.Value{Data: &messages.Value_Uint64Value{Uint64Value: uint64(value)}}
			case string:
				updateValue[k] = &messages.Value{Data: &messages.Value_StringValue{StringValue: value}}
			case []byte:
				updateValue[k] = &messages.Value{Data: &messages.Value_BytesValue{BytesValue: value}}
			}
		}
		a.lastWriteError = 0
		ctx.Request(ctx.Sender(), &messages.MsgWritePayment{
			Uid:     msg.Uid,
			Type:    msg.GetType(),
			Updates: updateValue,
		})
	case *messages.MsgWritePaymentError:
		a.lastWriteError = msg.Uid
		raw := msg.GetRaw()
		raw["mv"] = fmt.Sprintf("%d", a.mcard.VersionLayout())
		a.mcard.SetRawDataAfter(raw)
		a.mcard.SetError(msg.Error)
		a.fmachine.Event(eError, NewErrorScreen("error de escritura", "vuelva a ubicar la tarjeta"))
		logs.LogError.Printf("error de escritura, uid: %d, err: %s", msg.Uid, msg.Error)

	case *messages.MsgWritePaymentResponse:
		// ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: fmt.Sprintf("$%.02f", float32(a.mcard["newSaldo"].(int32)))})
		a.fmachine.Event(eCardValidated, a.mcard.Balance())

		raw := msg.GetRaw()
		if raw != nil {
			raw["mv"] = fmt.Sprintf("%d", a.mcard.VersionLayout())
		}

		a.mcard.SetRawDataAfter(raw)

		// if !ok {
		// 	logs.LogError.Println("seq is not INT")
		// }
		cardid := a.mcard.ID()
		uid := msg.Uid
		coord := ""
		if a.pidGps != nil {
			res, err := ctx.RequestFuture(a.pidGps, &gps.MsgGetGps{}, 10*time.Millisecond).Result()
			if err != nil {
				logs.LogWarn.Println("get fare err: %w", err)
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
		logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
		id, err := uuid.NewUUID()
		if err != nil {
			logs.LogError.Printf("uuid err: %s", err)
			break
		}
		uso := &usostransporte.UsoTransporte{
			ID:                    id.String(),
			DeviceID:              a.deviceID,
			PaymentMediumTypeCode: msg.Type,
			PaymentMediumId:       fmt.Sprintf("%d", cardid),
			MediumID:              fmt.Sprintf("%d", uid),
			FareCode:              int(a.mcard.FareID()),
			RawDataPrev:           a.mcard.RawDataBefore(),
			RawDataAfter:          a.mcard.RawDataAfter(),
			TransactionType:       "TRANSPORT_FARE_COLLECTION",
			Error: &usostransporte.Error{
				Name: "",
				Desc: "",
				Code: 0,
			},
			Coord: coord,
		}
		ctx.Send(a.pidUso, &usostransporte.MsgUso{
			Data: uso,
		})
		a.mcard = nil
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
		v := fmt.Sprintf(urlQr, a.params.PaymentItinerary, a.deviceID, msg.Value)
		ctx.Send(a.pidGraph, &graph.MsgQrValue{Value: v})
	}
}
