package mplus

import (
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type mplus struct {
	ttype         string
	uid           uint64
	id            uint
	historical    []payment.Historical
	balance       int
	profileID     uint
	pmr           bool
	ac            uint
	recharged     []payment.HistoricalRecharge
	consecutive   uint
	versionLayout uint
	lock          bool
	dateValidity  time.Time
	rawDataBefore interface{}
	rawDataAfter  interface{}
	actualMap     map[string]interface{}
	updateMap     map[string]interface{}
	fareID        uint
	err           string
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
func (p *mplus) UID() uint64 {
	return p.uid
}
func (p *mplus) ID() uint {
	return p.id
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
func (p *mplus) SetProfile(profile uint) {
	p.profileID = profile
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[PERFIL] = profile
}
func (p *mplus) SetLock() {
	p.lock = true
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[BLOQUEO] = 1
}

func (p *mplus) SetError(err string) {
	p.err = err
}

func (p *mplus) Error() string {
	return p.err
}
