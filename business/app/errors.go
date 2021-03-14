package app

import (
	"errors"
	"fmt"
)

var BalanceError = errors.New("Balance error")

type BalanceErrorValue struct {
	Balance float64
	Cost    float64
}

func (e *BalanceErrorValue) Unwrap() error {
	return BalanceError
}

func (e *BalanceErrorValue) Error() string {
	return fmt.Sprintf("saldo: %.2f", e.Balance)
}

var QrError = errors.New("QR error")

type QrErrorValue struct {
	Value string
}

func (e *QrErrorValue) Unwrap() error {
	return QrError
}

func (e *QrErrorValue) Error() string {
	return fmt.Sprintf("%s", e.Value)
}
