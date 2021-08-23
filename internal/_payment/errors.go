package payment

import (
	"errors"
	"fmt"
)

var ErrorBalance = errors.New("balance error")

type ErrorBalanceValue struct {
	Balance float64
	Cost    float64
}

func (e *ErrorBalanceValue) Unwrap() error {
	return ErrorBalance
}

func (e *ErrorBalanceValue) Error() string {
	return fmt.Sprintf("%.2f", e.Balance)
}
