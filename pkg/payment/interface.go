package payment

import (
	"time"
)

type Historical interface {
	Index() int
	FareID() uint
	TimeTransaction() time.Time
	ItineraryID() uint
	DeviceID() uint

	SetIndex(int)
	SetFareID(uint)
	SetTimeTransaction(time.Time)
	SetItineraryID(uint)
	SetDeviceID(uint)
}

type HistoricalRecharge interface {
	Index() int
	DeviceID() uint
	Value() int
	TimeTransaction() time.Time
	ConsecutiveID() uint
	TypeTransaction() uint

	SetIndex(int)
	SetDeviceID(uint)
	SetValue(int)
	SetTimeTransaction(time.Time)
	SetConsecutiveID(uint)
	SetTypeTransaction(uint)
}

type Payment interface {
	Type() string
	UID() uint64
	ID() uint
	Historical() []Historical
	Balance() int
	ProfileID() uint
	PMR() bool
	AC() uint
	Recharged() []HistoricalRecharge
	Consecutive() uint
	VersionLayout() uint
	Lock() bool
	Data() map[string]interface{}
	Updates() map[string]interface{}
	RawDataBefore() interface{}
	RawDataAfter() interface{}
	FareID() uint

	ApplyFare(data interface{}) (interface{}, error)
	AddRecharge(value int, deviceID, typeT, consecutive uint)
	AddBalance(value int) error
	SetProfile(uint)
	SetRawDataBefore(interface{})
	SetRawDataAfter(interface{})
	// IncConsecutive()
	SetError(err string)
	Error() string
	SetLock()
}
