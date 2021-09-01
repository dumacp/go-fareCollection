package recharge

import "time"

type RechargeQR struct {
	TID   int    `json:"tid"`
	Value int    `json:"v"`
	Exp   int64  `json:"e"`
	Date  int64  `json:"t"`
	PID   string `json:"pid"`
}

type Recharge struct {
	Value    int           `json:"v"`
	Exp      time.Duration `json:"e"`
	Date     time.Time     `json:"t"`
	PID      uint          `json:"pid"`
	TID      int           `json:"tid"`
	Updates  map[string]interface{}
	Seq      int
	DeviceID int
}
