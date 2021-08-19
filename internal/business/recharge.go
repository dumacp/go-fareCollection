package business

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func RechargeQR(paym payment.Payment, deviceID uint, seq uint,
	data *recharge.RechargeQR) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}
	exp := time.Unix(data.Exp/1000, (data.Exp%1000)*1_000_0000)
	if exp.Before(time.Now()) {
		return nil, nil
	}

	varSplit := strings.Split(data.PID, "-")
	id := 0
	if len(varSplit) > 0 {
		id, _ = strconv.Atoi(varSplit[len(varSplit)-1])
	}
	if paym.ID() != uint(id) {
		return nil, errors.New("tarjeta no coincide con al recarga")
	}
	if data.Value <= 0 {
		return nil, errors.New("valor invalido")
	}

	//TODO: where get this param
	ttype := 2
	paym.AddRecharge(data.Value, deviceID, uint(ttype), seq)

	return paym.Updates(), nil
}
