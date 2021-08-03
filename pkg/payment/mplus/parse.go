package mplus

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

type mplus struct {
	uid           uint64
	id            uint
	historical    []payment.Historical
	balance       int
	profileID     uint
	pmr           bool
	ac            uint
	recharged     []payment.HistoricalRecharge
	consecutive   uint
	versionLayout uint
	lock          bool
	dateValidity  time.Time
	rawDataBefore interface{}
	rawDataAfter  interface{}
	actualMap     map[string]interface{}
	updateMap     map[string]interface{}
}

// func extractInt(data interface{}) int {
// 	var result int
// 	switch v := data.(type) {
// 	case int:
// 		result = v
// 	case uint:
// 		result = int(v)
// 	default:
// 		return 0 //, errors.New("bad format")
// 	}
// 	return result //, nil
// }

func ParseToPayment(uid uint64, mapa map[string]interface{}) payment.Payment {

	// minorNumberInt32 := -2147483648

	histu := make(map[string]payment.Historical)
	histr := make(map[string]payment.HistoricalRecharge)

	m := &mplus{}
	m.uid = uid
	m.actualMap = mapa

	for k, value := range mapa {
		switch {
		case k == SaldoTarjeta:
			v, _ := value.(int)
			if v > m.balance {
				m.balance = v
			}
		case k == SaldoTarjetaBackup:
			v, _ := value.(int)
			if v > m.balance {
				m.balance = v
			}
		case k == PERFIL:
			m.profileID, _ = value.(uint)
		case k == ConsecutivoTarjeta:
			m.consecutive, _ = value.(uint)
		case k == BLOQUEO:
			v, _ := value.(int)
			if v > 0 {
				m.lock = true
			}
		case k == NUMEROTARJETA:
			m.id, _ = value.(uint)
		case k == VERSIONLAYOUT:
			m.versionLayout, _ = value.(uint)
		case k == PMR:
			v, _ := value.(uint)
			if v > 0 {
				m.pmr = true
			}
		case k == AC:
			m.ac, _ = value.(uint)
		case k == FechaValidezMonedero:
			v, _ := value.(uint)
			m.dateValidity = time.Unix(int64(v), int64(0))
		case strings.HasPrefix(k, HISTORICO_USO):
			re, err := regexp.Compile(fmt.Sprintf("(%s_.+)_([0-9])", HISTORICO_USO))
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
				histu[ind] = &historicalUse{index: indx}
			}
			switch key {
			case FechaTransaccion:
				v, _ := value.(uint)
				histu[ind].SetTimeTransaction(time.Unix(int64(v), int64(0)))
			case FareID:
				v, _ := value.(uint)
				histu[ind].SetFareID(v)
			case IDDispositivoUso:
				v, _ := value.(uint)
				histu[ind].SetDeviceID(v)
			case ItineraryID:
				v, _ := value.(uint)
				histu[ind].SetItineraryID(v)
			}
		case strings.HasPrefix(k, HISTORICO_RECARGA):
			re, err := regexp.Compile(fmt.Sprintf("(%s_.+)_([0-9])", HISTORICO_RECARGA))
			if err != nil {
				break
			}
			res := re.FindStringSubmatch(k)
			if len(res) <= 2 || len(res[1]) <= 0 {
				break
			}
			ind := res[2]
			key := res[1]
			if _, ok := histr[ind]; !ok {
				indx, _ := strconv.Atoi(ind)
				histr[ind] = &historicalRecharge{index: indx}
			}
			// fmt.Printf("hist: %v, %s, %v, %v\n", histr, ind, histr[ind], res)
			switch key {
			case FechaTransaccionRecarga:
				v, _ := value.(uint)
				histr[ind].SetTimeTransaction(time.Unix(int64(v), int64(0)))
			case TipoTransaccion:
				v, _ := value.(uint)
				histr[ind].SetTypeTransaction(v)
			case IDDispositivoRecarga:
				v, _ := value.(uint)
				histr[ind].SetDeviceID(v)
			case ValorTransaccionRecarga:
				v, _ := value.(int)
				histr[ind].SetValue(v)
			case ConsecutivoTransaccionRecarga:
				v, _ := value.(uint)
				histr[ind].SetConsecutiveID(v)
			}
		}
	}

	result := make([]payment.Historical, 0)
	for _, v := range histu {
		result = append(result, v)
	}

	sort.Slice(result,
		func(i, j int) bool {
			return result[i].TimeTransaction().Before(result[j].TimeTransaction())
		},
	)
	m.historical = result

	resultr := make([]payment.HistoricalRecharge, 0)
	for _, v := range histr {
		resultr = append(resultr, v)
	}

	sort.SliceStable(resultr,
		func(i, j int) bool {
			return resultr[i].TimeTransaction().Before(result[j].TimeTransaction())
		},
	)
	m.recharged = resultr

	return m
}

func (p *mplus) Data() map[string]interface{} {
	return p.actualMap
}
func (p *mplus) Updates() map[string]interface{} {
	return p.updateMap
}

func (p *mplus) UID() uint64 {
	return p.uid
}
func (p *mplus) ID() uint {
	return p.id
}

func (p *mplus) ProfileID() uint {
	return p.profileID
}
func (p *mplus) PMR() bool {
	return p.pmr
}
func (p *mplus) AC() uint {
	return p.ac
}

func (p *mplus) Consecutive() uint {
	return p.consecutive
}
func (p *mplus) VersionLayout() uint {
	return p.versionLayout
}
func (p *mplus) Lock() bool {
	return p.lock
}
func (p *mplus) RawDataBefore() interface{} {
	return nil
}
func (p *mplus) RawDataAfter() interface{} {
	return nil
}

func (p *mplus) SetProfile(profile uint) {
	p.profileID = profile
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[PERFIL] = profile
}

// func (p *mplus) IncConsecutive() {
// 	p.consecutive++
// }
func (p *mplus) SetLock() {
	p.lock = true
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[BLOQUEO] = 1
}
