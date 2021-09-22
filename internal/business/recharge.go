package business

import (
	"errors"
	"time"

	"github.com/dumacp/go-fareCollection/internal/recharge"
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func RechargeQR(paym payment.Payment,
	data *recharge.Recharge) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}

	if paym.PID() != data.PID {
		return nil, errors.New("tarjeta no coincide con la recarga")
	}
	if data.Value <= 0 {
		return nil, errors.New("valor invalido")
	}

	if data.Date.Add(data.Exp).Before(time.Now()) {
		return nil, errors.New("recarga expirada")
	}

	lasth := paym.Recharged()
	// tt := time.Now()
	if len(lasth) > 0 {
		// tt = lasth[0].TimeTransaction()
		for _, v := range lasth {
			if v.TimeTransaction().Before(time.Unix(969307158, 0)) {
				continue
			}
			if v.TimeTransaction().After(data.Date) {
				return nil, errors.New("recarga ya no es válida")
			}
			break
		}
	}
	// if tt.After(data.Date) {
	// 	return nil, errors.New("recarga ya no es válida")
	// }

	//TODO: where get this param
	ttype := 2
	if err := paym.AddRecharge(data.Value, uint(data.DeviceID), uint(ttype), uint(data.TID)); err != nil {
		return nil, err
	}

	if len(paym.Recharged()) > 0 {
		// logs.LogInfo.Printf("len hist re: %d", len(paym.Recharged()))
		paym.Recharged()[len(paym.Recharged())-1].SetRechargeProp("RechargeTokenId", data.TID)
		paym.Recharged()[len(paym.Recharged())-1].SetRechargeProp("Seq", data.Seq)
	}

	return paym.Updates(), nil
}
