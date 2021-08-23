package recharge

import "time"

type RechargeQR struct {
	TID   int    `json:"tid"`
	Value int    `json:"v"`
	Exp   int64  `json:"exp"`
	PID   string `json:"pid"`
}

type Recharge struct {
	Value    int       `json:"v"`
	Exp      time.Time `json:"exp"`
	PID      uint      `json:"pid"`
	TID      int       `json:"tid"`
	Updates  map[string]interface{}
	Seq      int
	DeviceID int
}
