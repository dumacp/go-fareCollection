package token

import (
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type token struct {
	id              string
	ttype           string
	pin             int
	pid             uint
	fid             uint
	exp             time.Time
	coord           string
	data            map[string]interface{}
	rawDataBefore   interface{}
	rawDataAfter    interface{}
	historical      []payment.Historical
	lock            bool
	lockReason      string
	lockList        string
	lockListVersion float32
	err             string
}

func (t *token) Type() string {
	return t.ttype
}

func (t *token) ID() string {
	return t.id
}

func (t *token) MID() uint64 {
	return uint64(t.pid)
}

func (t *token) PID() uint {
	return t.pid
}

func (t *token) ProfileID() uint {
	return 0
}

func (t *token) PMR() bool {
	return false
}

func (t *token) AC() uint {
	return 0
}

func (t *token) Recharged() []payment.HistoricalRecharge {
	return nil
}

func (t *token) Consecutive() uint {
	return 0
}

func (t *token) VersionLayout() uint {
	return 0
}

func (t *token) Lock() bool {
	return t.lock
}
func (t *token) LockReason() string {
	return t.lockReason
}
func (t *token) LockList() string {
	return t.lockList
}
func (t *token) LockListVersion() float32 {
	return t.lockListVersion
}

func (t *token) Data() map[string]interface{} {
	return t.data
}

func (t *token) Updates() map[string]interface{} {
	return nil
}

func (t *token) AddRecharge(value int, deviceID uint, typeT uint, consecutive uint) error {
	return nil
}

func (t *token) AddBalance(value int) error {
	return nil
}

func (t *token) SetProfile(_ uint) {}

func (t *token) SetLock(reason, listCode string, listVersion float32) {
	t.lockReason = reason
	t.lockList = listCode
	t.lockListVersion = listVersion
}

func (t *token) SetCoord(data string) {
	t.coord = data
}

func (t *token) Coord() string {
	return t.coord
}

func (t *token) SetError(err string) {
	t.err = err
}

func (t *token) Error() string {
	return t.err
}
