package app

import (
	"errors"
	"strings"
)

var ErrorQR = errors.New("QR error")
var ErrorScreen = errors.New("screen error")

type ErrorQRValue struct {
	Value string
}

func (e *ErrorQRValue) Unwrap() error {
	return ErrorQR
}

func (e *ErrorQRValue) Error() string {
	return e.Value
}

type ErrorShowInScreen struct {
	Value []string
}

func NewErrorScreen(data ...string) *ErrorShowInScreen {
	e := &ErrorShowInScreen{}
	e.Value = make([]string, 0)

	e.Value = append(e.Value, data...)

	return e
}

func (e *ErrorShowInScreen) Unwrap() error {
	return ErrorQR
}

func (e *ErrorShowInScreen) Error() string {
	return strings.Join(e.Value, ", ")
}
