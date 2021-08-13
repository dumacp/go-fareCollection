package token

import (
	"strconv"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func ParseToPayment(uid uint64, mapa map[string]interface{}) payment.Payment {
	t := &token{}
	t.ttype = "ENDUSER_QR"
	for k, value := range mapa {
		switch k {
		case "pid":
			t.ttype, _ = value.(string)
		case "pin":
			v, _ := value.(string)
			t.pin, _ = strconv.Atoi(v)
		}
	}
	return t
}
