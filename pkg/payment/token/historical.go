package token

import (
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type historical struct {
	fareID          uint
	timeTransaction time.Time
	itineraryID     uint
	deviceID        uint
}

func (t *token) Historical() []payment.Historical {
	hs := make([]payment.Historical, 0)
	hs = append(hs, t.historical...)
	return hs
}

func (t *historical) Index() int {
	return 0
}

func (t *historical) FareID() uint {
	return t.fareID
}

func (t *historical) TimeTransaction() time.Time {
	return t.timeTransaction
}

func (t *historical) ItineraryID() uint {
	return t.itineraryID
}

func (t *historical) DeviceID() uint {
	return t.deviceID
}

func (t *historical) SetIndex(_ int) {}

func (t *historical) SetFareID(fareID uint) {
	t.fareID = fareID
}

func (t *historical) SetTimeTransaction(date time.Time) {
	t.timeTransaction = date
}

func (t *historical) SetItineraryID(iti uint) {
	t.itineraryID = iti
}

func (t *historical) SetDeviceID(dev uint) {
	t.deviceID = dev
}
