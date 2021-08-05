package mplus

import (
	"fmt"
	"sort"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func (p *mplus) AddBalance(value int, deviceID, fareID, itineraryID uint) error {
	if p.balance <= 0 && value < 0 {
		return &payment.ErrorBalanceValue{Balance: float64(p.balance), Cost: float64(value)}
	}
	p.balance += value
	saldoTarjeta, _ := p.actualMap[SaldoTarjeta].(int)
	saldoTarjetaBackup, _ := p.actualMap[SaldoTarjetaBackup].(int)
	if value <= 0 {
		if saldoTarjeta > saldoTarjetaBackup {
			diff := saldoTarjeta - saldoTarjetaBackup
			saldoTarjeta += (value - diff)
		} else {
			diff := saldoTarjetaBackup - saldoTarjeta
			saldoTarjetaBackup += (value - diff)
		}
	} else {
		if saldoTarjeta > saldoTarjetaBackup {
			diff := saldoTarjeta - saldoTarjetaBackup
			if diff >= value {
				saldoTarjetaBackup += value
			} else {
				saldoTarjeta += (value - diff)
				saldoTarjetaBackup += value
			}
		} else {
			diff := saldoTarjetaBackup - saldoTarjeta
			if diff >= value {
				saldoTarjeta += value
			} else {
				saldoTarjetaBackup += (value - diff)
				saldoTarjeta += value
			}
		}
	}
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
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
	p.updateMap[fmt.Sprintf("%s_%d", FechaTransaccion, h.Index())] = uint(h.TimeTransaction().Unix())
	p.updateMap[fmt.Sprintf("%s_%d", FareID, h.Index())] = h.FareID()
	p.updateMap[fmt.Sprintf("%s_%d", ItineraryID, h.Index())] = h.ItineraryID()

	//Consecutive is a VALUE (int) in tag
	p.updateMap[ConsecutivoTarjeta] = int(p.consecutive + 1)

	sort.Slice(p.historical,
		func(i, j int) bool {
			return p.historical[i].TimeTransaction().Before(p.historical[j].TimeTransaction())
		},
	)

	return nil

}

func (p *mplus) Historical() []payment.Historical {
	hs := make([]payment.Historical, 0)
	hs = append(hs, p.historical...)
	return hs
}
func (p *mplus) Balance() int {
	return p.balance
}
