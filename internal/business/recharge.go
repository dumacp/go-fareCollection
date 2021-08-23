package business

import (
	"errors"
	"time"

	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/dumacp/go-logs/pkg/logs"
)

func RechargeQR(paym payment.Payment,
	data *recharge.Recharge) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}
	if data.Exp.Before(time.Now()) {
		return nil, nil
	}
	if paym.PID() != data.PID {
		return nil, errors.New("tarjeta no coincide con la recarga")
	}
	if data.Value <= 0 {
		return nil, errors.New("valor invalido")
	}

	//TODO: where get this param
	ttype := 2
	if err := paym.AddRecharge(data.Value, uint(data.DeviceID), uint(ttype), uint(data.Seq)); err != nil {
		return nil, err
	}

	if len(paym.Recharged()) > 0 {
		logs.LogInfo.Printf("len hist re: %d", len(paym.Recharged()))
		paym.Recharged()[len(paym.Recharged())-1].SetRechargeProp("RechargeTokenId", data.TID)
	}

	return paym.Updates(), nil
}
