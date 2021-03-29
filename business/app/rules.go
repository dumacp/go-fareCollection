package app

import (
	"errors"
	"fmt"
)

const (
	cost = 2400
)

func ValidationTag(tag map[string]interface{}) (map[string]interface{}, error) {

	saldo1, ok := tag["saldo"]
	if !ok {
		return nil, errors.New("saldo field is empty")
	}
	saldo2, ok := tag["saldoBackup"]
	if !ok {
		return nil, errors.New("saldoBk field is empty")
	}

	saldo := saldo1.(int32)
	if saldo > saldo2.(int32) {
		saldo = saldo2.(int32)
	}

	if saldo < cost {
		return nil, fmt.Errorf("error: %w", &ErrorBalanceValue{Balance: float64(saldo)})
	}
	newTag := make(map[string]interface{})
	if saldo1.(int32) > saldo2.(int32) {
		newTag["saldo"] = saldo1.(int32) - (saldo1.(int32) - saldo2.(int32)) - cost
		tag["newSaldo"] = saldo1.(int32) - cost
	} else {
		newTag["saldoBackup"] = saldo2.(int32) - (saldo2.(int32) - saldo1.(int32)) - cost
		tag["newSaldo"] = saldo2.(int32) - cost
	}

	v := tag["seq"].(int32) + 1
	newTag["seq"] = v
	return newTag, nil
}
func ValidationQr(tag map[string]interface{}) error {
	return nil
}
