package mplus

import (
	"errors"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type historicalUse struct {
	Index           int
	FareID          int
	TimeTransaction time.Time
	ItineraryID     int
	DeviceID        int
}
type historicalRecharged struct {
	// FareID          int
	TimeTransaction time.Time
	RechargedID     int
	TypeTransaction int
	// ItineraryID     int
	Value    int
	DeviceID int
}

type mplus struct {
	UID           int
	ID            int
	Historical    []*historicalUse
	Balance       int
	ProfileID     int
	PMR           bool
	AC            int
	Recharged     []*historicalRecharged
	Consecutive   int
	VersionLayout int
	Lock          bool
	DateVaslidity time.Time
	RawDataBefore interface{}
	RawDataAfter  interface{}
}

func ParseToPayment(uid int, mapa map[string]interface{}) payment.Payment {

	// minorNumberInt32 := -2147483648

	histu := make(map[string]*historicalUse)
	histr := make(map[string]*historicalRecharged)

	extractInt := func(data interface{}) (int, error) {
		var result int
		switch v := data.(type) {
		case int:
			result = v
		case uint:
			result = int(v)
		case uint64:
			result = int(v)
		case int64:
			result = int(v)
		default:
			return 0, errors.New("bad format")
		}
		return result, nil
	}

	// extractString := func(data interface{}) (string, error) {
	// 	var result string
	// 	switch v := data.(type) {
	// 	case string:
	// 		result = v
	// 	default:
	// 		return "", errors.New("bad format")
	// 	}
	// 	return result, nil
	// }

	m := &mplus{}
	m.UID = uid
	for k, value := range mapa {
		switch {
		case k == SaldoTarjeta:
			if v, err := extractInt(value); err == nil {
				if m.Balance > v {
					m.Balance = v
				}
			}
		case k == SaldoTarjetaBackup:
			if v, err := extractInt(value); err == nil {
				if m.Balance > v {
					m.Balance = v
				}
			}
		case k == PERFIL:
			m.ProfileID, _ = extractInt(value)
		case k == ConsecutivoTarjeta:
			m.Consecutive, _ = extractInt(value)
		case k == BLOQUEO:
			v, _ := extractInt(value)
			if v > 0 {
				m.Lock = true
			}
		case k == NUMEROTARJETA:
			m.ID, _ = extractInt(value)
		case k == VERSIONLAYOUT:
			m.VersionLayout, _ = extractInt(value)
		case k == PMR:
			v, _ := extractInt(value)
			if v > 0 {
				m.PMR = true
			}
		case k == AC:
			m.AC, _ = extractInt(value)
		case k == FechaValidezMonedero:
			v, _ := extractInt(value)
			m.DateVaslidity = time.Unix(int64(v), int64(0))
		case strings.HasPrefix(k, "HISTU"):
			re, err := regexp.Compile(`HISTU_(.+)_([0-9])`)
			if err != nil {
				break
			}
			res := re.FindStringSubmatch(k)
			if len(res) <= 2 || len(res[1]) <= 0 {
				break
			}
			ind := res[2]
			key := res[1]
			if _, ok := histu[ind]; !ok {
				indx, _ := strconv.Atoi(ind)
				histu[ind] = &historicalUse{Index: indx}
			}
			switch key {
			case "FT":
				v, _ := extractInt(value)
				histu[ind].TimeTransaction = time.Unix(int64(v), int64(0))
			case "FID":
				histu[ind].FareID, _ = extractInt(value)
			case "IDV":
				histu[ind].DeviceID, _ = extractInt(value)
			case "ITI":
				histu[ind].ItineraryID, _ = extractInt(value)
			}
		case strings.HasPrefix(k, "HISTR"):
			re, err := regexp.Compile(`HISTR_(.+)_([0-9])`)
			if err != nil {
				break
			}
			res := re.FindStringSubmatch(k)
			if len(res) <= 2 || len(res[1]) <= 0 {
				break
			}
			ind := res[2]
			key := res[1]
			if _, ok := histu[ind]; !ok {
				indx, _ := strconv.Atoi(ind)
				histu[ind] = &historicalUse{Index: indx}
			}
			switch key {
			case "FT":
				v, _ := extractInt(value)
				histr[ind].TimeTransaction = time.Unix(int64(v), int64(0))
			case "TT":
				histr[ind].TypeTransaction, _ = extractInt(value)
			case "IDV":
				histr[ind].DeviceID, _ = extractInt(value)
			case "VT":
				histr[ind].Value, _ = extractInt(value)
			case "CT":
				histr[ind].RechargedID, _ = extractInt(value)
			}
		}
	}

	result := make([]*historicalUse, 0)
	for _, v := range histu {
		result = append(result, v)
	}

	sort.SliceStable(result,
		func(i, j int) bool {
			return result[i].TimeTransaction.Before(result[j].TimeTransaction)
		},
	)
	m.Historical = result

	resultr := make([]*historicalRecharged, 0)
	for _, v := range histr {
		resultr = append(resultr, v)
	}

	sort.SliceStable(resultr,
		func(i, j int) bool {
			return resultr[i].TimeTransaction.Before(result[j].TimeTransaction)
		},
	)
	m.Recharged = resultr

	return nil

}
