package payment

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
