package payment

import (
	"time"
)

type Historical interface {
	FareID() int
	TimeTransaction() time.Time
	ItineraryID() int
	DeviceID() int
}

type HistoricalRecharge interface {
	DeviceID() int
	Value() int
	time() time.Time
	ConsecutiveID() int
	TypeTransaction() int
}

type Payment interface {
	UID() int
	ID() int
	Historical() []Historical
	Balance() int
	ProfileID() int
	PMR() bool
	AC() int
	Recharged() []HistoricalRecharge
	Consecutive() int
	VersionLayout() int
	Lock() bool
	RawDataBefore() interface{}
	RawDataAfter() interface{}

	AddRecharge(value, deviceID, typeT, consecutive int)
	AddBalance(value, deviceID, fareID, itineraryID int)
	SetProfile(int)
	IncConsecutive()
	SetLock()
}
