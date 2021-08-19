package recharge

type RechargeQR struct {
	Pocket uint   `json:"p"`
	Value  int    `json:"v"`
	Exp    int64  `json:"exp"`
	PID    string `json:"pid"`
}
