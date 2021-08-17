package token

import (
	"strconv"
	"strings"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func ParseToPayment(uid uint64, mapa map[string]interface{}) payment.Payment {
	t := &token{}
	t.ttype = "ENDUSER_QR"
	t.data = mapa
	for k, value := range mapa {
		switch k {
		case "pid":
			switch v := value.(type) {
			case int:
				t.id = uint(v)
			case string:
				varsplit := strings.Split(v, "-")
				if len(varsplit) > 0 {
					num, _ := strconv.Atoi(varsplit[len(varsplit)-1])
					t.id = uint(num)
				}
			}
		case "pin":
			v, _ := value.(string)
			t.pin, _ = strconv.Atoi(v)
		}
	}
	return t
}
