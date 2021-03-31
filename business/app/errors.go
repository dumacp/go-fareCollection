package app

import (
	"errors"
	"fmt"
)

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
