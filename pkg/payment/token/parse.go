package token

import (
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
	"github.com/google/uuid"
)

func ParseToPayment(uid uint64, ttype string, mapa map[string]interface{}) payment.Payment {
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
	t.ttype = ttype
	t.data = mapa
	t.historical = make([]payment.Historical, 0)
	for k, value := range mapa {
		switch k {
		case "pid":
			switch v := value.(type) {
			case int:
				t.pid = uint(v)
			case uint:
				t.pid = uint(v)
			case int64:
				t.pid = uint(v)
			case string:
				varsplit := strings.Split(v, "-")
				if len(varsplit) > 0 {
					num, _ := strconv.Atoi(varsplit[len(varsplit)-1])
					t.pid = uint(num)
				}
			}
		case "fid":
			switch v := value.(type) {
			case int:
				t.fid = uint(v)
			case uint:
				t.fid = uint(v)
			case int64:
				t.fid = uint(v)
			}
		case "e":
			exp := 0
			switch v := value.(type) {
			case int:
				exp = int(v)
			case uint:
				exp = int(v)
			case int64:
				exp = int(v)
			default:
			}
			t.exp = time.Duration(exp) * time.Second
		case "t":
			date := int64(0)
			switch v := value.(type) {
			case int:
				date = int64(v)
			case uint:
				date = int64(v)
			case int64:
				date = int64(v)
			default:
			}
			t.date = time.Unix(date, 0)
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
