package token

import (
	"errors"
	"time"

	"github.com/dumacp/go-fareCollection/internal/qr"
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func (t *token) ApplyFare(data interface{}) (interface{}, error) {

	switch t.ttype {
	case qr.EQPM:

		switch value := data.(type) {
		case []int:
			for _, v := range value {
				if t.pin == v {
					if len(t.historical) <= 0 {
						t.historical = make([]payment.Historical, 0)
					}
					h0 := &historical{
						fareID:          0,
						timeTransaction: time.Now(),
						itineraryID:     0,
						deviceID:        0,
					}
					t.historical = append(t.historical, h0)

					return v, nil
				}
			}
			return 0, errors.New("pin no es válido")
		}

		return 0, errors.New("pin no es válido")
	case qr.AQPM:
		if t.exp.Before(time.Now()) {
			return 0, errors.New("exp no es válido")
		}
		if len(t.historical) <= 0 {
			t.historical = make([]payment.Historical, 0)
		}
		h0 := &historical{
			fareID:          t.fid,
			timeTransaction: time.Now(),
			itineraryID:     0,
			deviceID:        0,
		}
		t.historical = append(t.historical, h0)

		return 100, nil
	default:
		return 0, errors.New("qr no es válido")
	}
}

func (t *token) Balance() int {
	return int(t.pid)
}

func (t *token) FareID() uint {
	return t.fid
}
