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

	// extractInt := func(data interface{}) int {
	// 	var result int
	// 	switch v := data.(type) {
	// 	case int:
	// 		result = v
	// 	case uint:
	// 		result = int(v)
	// 	case uint64:
	// 		result = int(v)
	// 	case int64:
	// 		result = int(v)
	// 	default:
	// 		return 0 //, errors.New("bad format")
	// 	}
	// 	return result //, nil
	// }

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
	m.uid = uid
	m.actualMap = mapa
	for k, value := range mapa {
		switch {
		case k == SaldoTarjeta:
			m.balance, _ = value.(int)
		case k == SaldoTarjetaBackup:
			m.balance, _ = value.(int)
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
			if _, ok := histu[ind]; !ok {
				indx, _ := strconv.Atoi(ind)
				histu[ind] = &historicalUse{index: indx}
			}
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

	sort.SliceStable(result,
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

func (p *mplus) UID() uint64 {
	return p.uid
}
func (p *mplus) ID() uint {
	return p.id
}
func (p *mplus) Historical() []payment.Historical {
	return nil
}
func (p *mplus) Balance() int {
	return p.balance
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
func (p *mplus) Recharged() []payment.HistoricalRecharge {
	hs := make([]payment.HistoricalRecharge, 0)
	for _, v := range p.recharged {
		hs = append(hs, v)
	}
	return hs
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

func (p *mplus) AddRecharge(value int, deviceID, typeT, consecutive uint) {
	if len(p.recharged) <= 0 {
		p.recharged = make([]payment.HistoricalRecharge, 0)
		p.recharged = append(p.recharged, &historicalRecharge{})

	}
	p.recharged[0].SetValue(value)
	p.recharged[0].SetDeviceID(deviceID)
	p.recharged[0].SetTypeTransaction(typeT)
	p.recharged[0].SetTimeTransaction(time.Now())
	p.recharged[0].SetConsecutiveID(consecutive)
}
func (p *mplus) AddBalance(value int, deviceID, fareID, itineraryID uint) {
	p.balance += value
	saldoTarjeta, _ := p.actualMap[SaldoTarjeta].(int)
	saldoTarjetaBackup, _ := p.actualMap[SaldoTarjetaBackup].(int)
	if value > 0 {
		if saldoTarjeta > saldoTarjetaBackup {
			diff := saldoTarjeta - saldoTarjetaBackup
			if diff > value {
				saldoTarjetaBackup += value
			} else {
				saldoTarjeta += value - diff
			}
		} else {
			diff := saldoTarjetaBackup - saldoTarjeta
			if diff > value {
				saldoTarjeta += value
			} else {
				saldoTarjetaBackup += value - diff
			}
		}
	} else {
		if saldoTarjeta > saldoTarjetaBackup {
			saldoTarjeta += value
		} else {
			saldoTarjetaBackup += value
		}
	}
	p.updateMap[SaldoTarjeta] = saldoTarjeta
	p.updateMap[SaldoTarjetaBackup] = saldoTarjetaBackup
	if len(p.historical) <= 0 {
		p.historical = make([]payment.Historical, 0)
		p.historical = append(p.historical, &historicalUse{})
		p.historical[0].SetIndex(1)

	}
	p.historical[0].SetDeviceID(deviceID)
	p.historical[0].SetFareID(fareID)
	p.historical[0].SetItineraryID(itineraryID)
	p.historical[0].SetTimeTransaction(time.Now())

	h := p.historical[0]
	p.updateMap[fmt.Sprintf("%s_%d", IDDispositivoUso, h.Index())] = h.DeviceID()
	p.updateMap[fmt.Sprintf("%s_%d", FechaTransaccion, h.Index())] = h.TimeTransaction()
	p.updateMap[fmt.Sprintf("%s_%d", FareID, h.Index())] = h.FareID()
	p.updateMap[fmt.Sprintf("%s_%d", ItineraryID, h.Index())] = h.ItineraryID()

}
func (p *mplus) SetProfile(profile uint) {
	p.profileID = profile
}
func (p *mplus) IncConsecutive() {
	p.consecutive++
}
func (p *mplus) SetLock() {
	p.lock = true
}
