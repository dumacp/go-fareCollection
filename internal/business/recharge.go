package business

import (
	"errors"
	"time"

	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func RechargeQR(paym payment.Payment, deviceID uint, seq uint,
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
	if err := paym.AddRecharge(data.Value, deviceID, uint(ttype), seq); err != nil {
		return nil, err
	}

	return paym.Updates(), nil
}
