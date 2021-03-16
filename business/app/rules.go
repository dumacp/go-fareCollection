package app

import (
	"errors"
	"fmt"
)

func ValidationTag(tag map[string]interface{}) error {

	saldo1, ok := tag["saldo"]
	if !ok {
		return errors.New("saldo field is empty")
	}
	saldo2, ok := tag["saldoBackup"]
	if !ok {
		return errors.New("saldoBk field is empty")
	}

	saldo := saldo1.(float64)
	if saldo > saldo2.(float64) {
		saldo = saldo2.(float64)
	}

	if saldo < 2000 {
		return fmt.Errorf("error: %w", &ErrorBalanceValue{Balance: saldo})
	}
	return nil
}
func ValidationQr(tag map[string]interface{}) error {
	return nil
}
