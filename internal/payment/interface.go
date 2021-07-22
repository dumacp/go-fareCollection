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

type Payment interface {
	UID() int
	ID() int
	Historical() []Historical
	Balance() int
	ProfileID() int
	PMR() bool
	AC() int
	Recharged() int
	Consecutive() int
	VersionLayout() int
	Lock() bool
	RawDataBefore() interface{}
	RawDataAfter() interface{}

	AddRecharge(int)
	AddBalance(int)
	WriteProfile(int)
	IncConsecutive()
	SetLock()
	AddHistorical(fareID, itineraryID, deviceID int, date time.Time)
}
