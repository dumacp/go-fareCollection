package business

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/fare"
	"github.com/dumacp/go-fareCollection/internal/lists"
	"github.com/dumacp/go-fareCollection/internal/parameters"
	"github.com/dumacp/go-fareCollection/pkg/messages"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-fareCollection/pkg/payment/mplus"
	"github.com/dumacp/go-fareCollection/pkg/payment/token"
	"github.com/dumacp/go-logs/pkg/logs"
)

func ParsePayment(msg *messages.MsgPayment) (payment.Payment, error) {
	/**/
	jsonprint, err := json.MarshalIndent(msg.Data, "", "  ")
	if err != nil {
		logs.LogError.Println(err)
	}
	logs.LogBuild.Printf("tag read: %s", jsonprint)

	/**/
	var paym payment.Payment
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
	// v, err := payment.ValidationTag(lastMcard, 1028, 1290)

	switch msg.GetType() {
	case "MIFARE_PLUS_EV2_4K":
		paym = mplus.ParseToPayment(msg.Uid, msg.GetType(), mcard)
	case "ENDUSER_QR":
		paym = token.ParseToPayment(msg.Uid, mcard)
	}

	raw := msg.GetRaw()
	raw["mv"] = fmt.Sprintf("%d", paym.VersionLayout())
	paym.SetRawDataBefore(raw)
	return paym, nil

}

func VerifyListRestrictive(ctx actor.Context, pidList *actor.PID, paym payment.Payment,
	listRestrictive map[string]string) (bool, error) {
	for list, code := range listRestrictive {
		if code != paym.Type() {
			continue
		}
		resList, err := ctx.RequestFuture(pidList, &lists.MsgVerifyInList{
			ListID: list,
			ID:     []int64{int64(paym.PID())},
		}, 60*time.Millisecond).Result()
		if err != nil {
			return false, fmt.Errorf("get restrictive list err: %w", err)

		}
		switch v := resList.(type) {
		case *lists.MsgVerifyInListResponse:
			if len(v.ID) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}

func CalcUpdatesWithFare(ctx actor.Context, pidFare *actor.PID, deviceID int,
	paym payment.Payment, params *parameters.Parameters) (map[string]interface{}, error) {
	switch paym.Type() {
	case "MIFARE_PLUS_EV2_4K":
		lastFares := make(map[int64]int)
		hs := paym.Historical()
		for _, v := range hs {
			timestamp := v.TimeTransaction().Unix()
			fareid := v.FareID()
			lastFares[timestamp] = int(fareid)
		}
		getFare := &fare.MsgGetFare{
			LastFarePolicies: lastFares,
			ProfileID:        int(paym.ProfileID()),
			// ItineraryID:      157,
			// ModeID:           1,
			// RouteID:         77,
			FromItineraryID: int(hs[len(hs)-1].ItineraryID()),
		}
		if params == nil {
			return nil, errors.New("params is nil")
		}
		getFare.ModeID = int(params.PaymentMode)
		//TODO: get ids by QR
		getFare.RouteID = int(params.PaymentRoute)
		getFare.ItineraryID = int(params.PaymentItinerary)
		//TODO: farePID?
		if pidFare == nil {
			return nil, errors.New("pidFare not found")
		}
		resFare, err := ctx.RequestFuture(pidFare, getFare, 60*time.Millisecond).Result()
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
			fareData.DeviceID = deviceID
			fareData.ItineraryID = params.PaymentItinerary
		case *fare.MsgError:
			return nil, errors.New(res.Err)
		default:
			return nil, errors.New("fareID not found")
		}
		logs.LogInfo.Printf("fare calc: %+v", fareData)

		if _, err := paym.ApplyFare(fareData); err != nil {
			// time.Sleep(3 * time.Second)
			return nil, err
		}
		return paym.Updates(), nil
	}
	return nil, nil
}

func CalcUpdatesQR(
	paym payment.Payment, oldRand, newRand *int) (map[string]interface{}, error) {
	switch paym.Type() {
	case "ENDUSER_QR":
		if pin, err := paym.ApplyFare([]int{*newRand, *oldRand}); err != nil {
			return nil, err
		} else {
			for _, v := range []*int{newRand, oldRand} {
				if *v == pin {
					*v = -1
				}
			}
		}
	}
	return nil, nil
}

func ParseUpdatesToValuePayment(update map[string]interface{}) map[string]*messages.Value {

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
	return updateValue
}
