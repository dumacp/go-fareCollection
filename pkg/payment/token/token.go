package token

import "github.com/dumacp/go-fareCollection/pkg/payment"

type token struct {
	ttype         string
	pin           int
	id            uint
	data          map[string]interface{}
	rawDataBefore interface{}
	rawDataAfter  interface{}
}

func (t *token) Type() string {
	return t.ttype
}

func (t *token) UID() uint64 {
	return uint64(t.id)
}

func (t *token) ID() uint {
	return t.id
}

func (t *token) Historical() []payment.Historical {
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
	// panic("not implemented") // TODO: Implement
	return 0
}

func (t *token) Lock() bool {
	panic("not implemented") // TODO: Implement
}

func (t *token) Data() map[string]interface{} {
	return t.data
}

func (t *token) Updates() map[string]interface{} {
	return nil
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

func (t *token) SetError(err string) {
	panic("not implemented") // TODO: Implement
}

func (t *token) Error() string {
	panic("not implemented") // TODO: Implement
}
