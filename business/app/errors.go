package app

import (
	"errors"
	"fmt"
)

var ErrorBalance = errors.New("Balance error")

type ErrorBalanceValue struct {
	Balance float64
	Cost    float64
}

func (e *ErrorBalanceValue) Unwrap() error {
	return ErrorBalance
}

func (e *ErrorBalanceValue) Error() string {
	return fmt.Sprintf("saldo: %.2f", e.Balance)
}

var ErrorQR = errors.New("QR error")

type ErrorQRValue struct {
	Value string
}

func (e *ErrorQRValue) Unwrap() error {
	return ErrorQR
}

func (e *ErrorQRValue) Error() string {
	return fmt.Sprintf("%s", e.Value)
}
