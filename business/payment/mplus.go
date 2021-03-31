package payment

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// type Cardv1 interface {
// 	Name() string
// 	DocID() string
// 	CardID() int
// 	Perfil() int
// 	Version() int
// 	PMR() int
// 	AC() int
// 	Bloqueo() bool
// 	FechaBloque() time.Time
// 	Saldo() int
// 	saldoBackup() int
// 	Seq() int
// 	FechaRecarga() time.Time
// 	SeqRecarga() int
// 	ValorRecarga() int
// 	TipoRecarga()
// 	FechaValidez() time.Time
// }

const (
	cost = 2400
)

func ValidationTag(tag map[string]interface{}, ruta int) (map[string]interface{}, error) {

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

type hist struct {
	date       time.Time
	valor      int
	devID      int
	route      int
	perfil     int
	seq        int
	walletType int
}

func sortHistory(tag map[string]interface{}) ([]*hist, error) {

	re, err := regexp.Compile(`hits[0-9]`)
	if err != nil {
		return nil, err
	}
	for k, v := range tag {
		re.
	}
	return nil, nil
}
