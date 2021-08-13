package token

import "github.com/dumacp/go-fareCollection/pkg/payment"

type token struct {
	ttype         string
	pin           int
	rawDataBefore interface{}
	rawDataAfter  interface{}
}

func (t *token) Type() string {
	return t.ttype
}

func (t *token) UID() uint64 {
	panic("not implemented") // TODO: Implement
}

func (t *token) ID() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) Historical() []payment.Historical {
	panic("not implemented") // TODO: Implement
}

func (t *token) Balance() int {
	panic("not implemented") // TODO: Implement
}

func (t *token) ProfileID() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) PMR() bool {
	panic("not implemented") // TODO: Implement
}

func (t *token) AC() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) Recharged() []payment.HistoricalRecharge {
	panic("not implemented") // TODO: Implement
}

func (t *token) Consecutive() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) VersionLayout() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) Lock() bool {
	panic("not implemented") // TODO: Implement
}

func (t *token) Data() map[string]interface{} {
	panic("not implemented") // TODO: Implement
}

func (t *token) Updates() map[string]interface{} {
	panic("not implemented") // TODO: Implement
}

func (t *token) FareID() uint {
	panic("not implemented") // TODO: Implement
}

func (t *token) AddRecharge(value int, deviceID uint, typeT uint, consecutive uint) {
	panic("not implemented") // TODO: Implement
}

func (t *token) AddBalance(value int) error {
	panic("not implemented") // TODO: Implement
}

func (t *token) SetProfile(_ uint) {
	panic("not implemented") // TODO: Implement
}

func (t *token) SetLock() {
	panic("not implemented") // TODO: Implement
}
