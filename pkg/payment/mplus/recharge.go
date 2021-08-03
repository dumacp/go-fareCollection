package mplus

import (
	"fmt"
	"time"

	"github.com/dumacp/go-fareCollection/pkg/payment"
)

func (p *mplus) AddRecharge(value int, deviceID, typeT, consecutive uint) {
	if len(p.recharged) <= 0 {
		p.recharged = make([]payment.HistoricalRecharge, 0)
		p.recharged = append(p.recharged, &historicalRecharge{})
		p.recharged[0].SetIndex(1)

	}
	p.recharged[0].SetValue(value)
	p.recharged[0].SetDeviceID(deviceID)
	p.recharged[0].SetTypeTransaction(typeT)
	p.recharged[0].SetTimeTransaction(time.Now())
	p.recharged[0].SetConsecutiveID(consecutive)

	h := p.recharged[0]
	if p.updateMap == nil {
		p.updateMap = make(map[string]interface{})
	}
	p.updateMap[fmt.Sprintf("%s_%d", IDDispositivoRecarga, h.Index())] = h.DeviceID()
	p.updateMap[fmt.Sprintf("%s_%d", FechaTransaccionRecarga, h.Index())] = uint(h.TimeTransaction().Unix())
	p.updateMap[fmt.Sprintf("%s_%d", ConsecutivoTransaccionRecarga, h.Index())] = h.ConsecutiveID()
	p.updateMap[fmt.Sprintf("%s_%d", TipoTransaccion, h.Index())] = h.TypeTransaction()
	p.updateMap[fmt.Sprintf("%s_%d", ValorTransaccionRecarga, h.Index())] = h.Value()
}

func (p *mplus) Recharged() []payment.HistoricalRecharge {
	hs := make([]payment.HistoricalRecharge, 0)
	hs = append(hs, p.recharged...)
	return hs
}
