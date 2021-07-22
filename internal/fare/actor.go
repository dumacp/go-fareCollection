package fare

import (
	"errors"
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/itinerary"
	"github.com/dumacp/go-logs/pkg/logs"
)

type Actor struct {
	farePolicies map[int]*FareNode
	fareMap      FareMap
	itineraryMap itinerary.ItineraryMap
	pidItinerary *actor.PID
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in readerActor: %+v, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.farePolicies = make(map[int]*FareNode)
		ctx.Send(ctx.Self(), &MsgGetFarePolicies{})
	case *MsgGetFarePolicies:
		//TODO:
		//Get Fare Policies from platform

	case *MsgGetFare:
		if fare, err := func() (*FareNode, error) {
			q := &QueryFare{
				Time:            time.Now(),
				FromItineraryID: msg.FromItineraryID,
				ItineraryID:     msg.ItineraryID,
				RouteID:         msg.RouteID,
				ModeID:          msg.ModeID,
				LastFare:        make([]*FareNode, 0),
			}
			if msg.FromItineraryID > 0 {
				if a.itineraryMap == nil ||
					a.itineraryMap[msg.ItineraryID] == nil {
					logs.LogError.Printf("itinerary not found")
					if a.pidItinerary != nil {
						ctx.Request(a.pidItinerary, &itinerary.MsgGetMap{})
					}
				} else {
					q.FromModeID = a.itineraryMap[msg.ItineraryID].ModeID
					q.FromRouteID = a.itineraryMap[msg.ItineraryID].RouteID
				}
			}
			keySort := make([]int, 0)
			for k := range msg.LastFarePolicies {
				keySort = append(keySort, k)
			}
			sort.Ints(keySort)

			foundPlain := false
			for _, k := range keySort {
				fare := a.farePolicies[msg.LastFarePolicies[k]]
				if !foundPlain {
					if fare.Type == PLAIN {
						foundPlain = true
						q.LastFare = append(q.LastFare, fare)
						q.LastTimePlain = time.Unix(int64(k), 0)
					}
					continue
				}
				q.LastFare = append(q.LastFare, fare)
			}
			fare := a.fareMap.FindFare(q)
			if fare == nil {
				return nil, errors.New("fare Policy not found")
			}
			return fare, nil
		}(); err != nil {
			if ctx.Sender() != nil {
				ctx.Respond(&MsgError{
					Err: err.Error(),
				})
			}
		} else {
			if ctx.Sender() != nil {
				ctx.Respond(&MsgResponseFare{
					Fare:         fare.Fare,
					FarePolicyID: fare.ID,
				})
			}
		}

		// calculate(msg.LastFarePolicies, a.farePolicies)
	}
}
