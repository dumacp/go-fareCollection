package token

import (
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/google/uuid"
)

func ParseToPayment(uid uint64, mapa map[string]interface{}) payment.Payment {
	cid := make(chan string)
	go func() {
		defer close(cid)
		id, _ := uuid.NewUUID()
		select {
		case cid <- id.String():
		case <-time.After(300 * time.Millisecond):
		}
	}()
	t := &token{}
	t.ttype = "ENDUSER_QR"
	t.data = mapa
	for k, value := range mapa {
		switch k {
		case "pid":
			switch v := value.(type) {
			case int:
				t.pid = uint(v)
			case string:
				varsplit := strings.Split(v, "-")
				if len(varsplit) > 0 {
					num, _ := strconv.Atoi(varsplit[len(varsplit)-1])
					t.pid = uint(num)
				}
			}
		case "pin":
			v, _ := value.(string)
			t.pin, _ = strconv.Atoi(v)
		}
	}
	// id, _ := uuid.NewUUID()
	// t.id = id.String()
	for id := range cid {
		t.id = id
	}
	return t
}
