package mplus

import (
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type mplus struct {
	id              string
	ttype           string
	mid             uint64
	pid             uint
	historical      []payment.Historical
	balance         int
	profileID       uint
	pmr             bool
	ac              uint
	recharged       []payment.HistoricalRecharge
	consecutive     uint
	versionLayout   uint
	lock            bool
	dateValidity    time.Time
	rawDataBefore   interface{}
	rawDataAfter    interface{}
	actualMap       map[string]interface{}
	updateMap       map[string]interface{}
	fareID          uint
	lockReason      string
	lockList        string
	lockListVersion float32
	coord           string
	err             string
}

func (p *mplus) ID() string {
	return p.id
}
func (p *mplus) Type() string {
	return p.ttype
}
func (p *mplus) Data() map[string]interface{} {
	return p.actualMap
}
func (p *mplus) Updates() map[string]interface{} {
	return p.updateMap
}
func (p *mplus) MID() uint64 {
	return p.mid
}
func (p *mplus) PID() uint {
	return p.pid
}
func (p *mplus) ProfileID() uint {
	return p.profileID
}
func (p *mplus) PMR() bool {
	return p.pmr
}
func (p *mplus) AC() uint {
	return p.ac
}
func (p *mplus) Consecutive() uint {
	return p.consecutive
}
func (p *mplus) VersionLayout() uint {
	return p.versionLayout
}
func (p *mplus) Lock() bool {
	return p.lock
}
func (p *mplus) LockReason() string {
	return p.lockReason
}
func (p *mplus) LockList() string {
	return p.lockList
}
func (p *mplus) LockListVersion() float32 {
	return p.lockListVersion
}
func (p *mplus) SetProfile(profile uint) {
	p.profileID = profile
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[PERFIL] = profile
}
func (p *mplus) SetLock(reason, listCode string, listVersion float32) {
	p.lock = true
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[BLOQUEO] = 1
	p.lockList = listCode
	p.lockReason = reason
	p.lockListVersion = listVersion
}

func (p *mplus) SetCoord(data string) {
	p.coord = data
}

func (p *mplus) Coord() string {
	return p.coord
}

func (p *mplus) SetError(err string) {
	p.err = err
}

func (p *mplus) Error() string {
	return p.err
}
