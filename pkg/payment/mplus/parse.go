package mplus

import (
	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type historical struct {
	// FareID          int
	// TimeTransaction time.Time
	// ItineraryID     int
	// DeviceID        int
}

type mplus struct {
	UID           int
	ID            int
	Historical    []Historical
	Balance       int
	ProfileID     int
	PMR           bool
	AC            int
	Recharged     int
	Consecutive   int
	VersionLayout int
	Lock          bool
	RawDataBefore interface{}
	RawDataAfter  interface{}
}

func ParseToPayment(uid int, mapa map[string]interface{}) payment.Payment {

	m := &mplus{}
	m.UID = uid
	for k, v := range mapa {
		switch k {
		case SaldoTarjeta:
			if m.Balance > 0 && v > 
		}
	}

	return nil

}
