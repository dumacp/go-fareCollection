package app

import (
	"fmt"

	"github.com/dumacp/go-fareCollection/internal/usostransporte"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/google/uuid"
)

func (a *Actor) writeerrorverify() {
	if a.lastWriteError <= 0 {
		return
	}
	if len(a.paym) > 0 && a.paym[a.lastWriteError] != nil {
		paym := a.paym[a.lastWriteError]
		logs.LogBuild.Printf("nodeID: [% X]", uuid.NodeID())
		lasth := paym.Historical()

		tt := int64(0)
		if len(lasth) > 0 {
			tt = lasth[len(lasth)-1].TimeTransaction().UnixNano() / 1_000_000
		}
		uso := &usostransporte.UsoTransporte{
			ID:                    paym.ID(),
			DeviceID:              a.deviceID,
			PaymentMediumTypeCode: paym.Type(),
			PaymentMediumId:       fmt.Sprintf("%d", paym.PID()),
			MediumID:              fmt.Sprintf("%d", paym.MID()),
			FareCode:              int(paym.FareID()),
			RawDataPrev:           paym.RawDataBefore(),
			RawDataAfter:          paym.RawDataAfter(),
			TransactionType:       "TRANSPORT_FARE_COLLECTION",
			TransactionTime:       tt,
			Error: &usostransporte.Error{
				Name: "write error",
				Desc: paym.Error(),
				Code: 1,
			},
			Coord: paym.Coord(),
		}
		a.ctx.Send(a.pidUso, uso)
	}
}
