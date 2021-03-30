package usotransport

import "time"

type UsoTransport struct {
	ID           string
	TimeStamp    time.Time
	AccountID    string
	PaymentID    string
	Cost         int
	RouteID      int
	FarePolicyID int
	Balance      int
	History      []*UsoTransport
}

func NewUso(accountID, paymentID, routeID int) (*UsoTransport, error) {
	return nil, nil
}
