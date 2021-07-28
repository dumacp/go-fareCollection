package itinerary

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
)

type Actor struct {
	quit         chan int
	httpClient   *http.Client
	userHttp     string
	passHttp     string
	urlMode      string
	urlRoute     string
	urlItinerary string
	itineraryMap ItineraryMap
}

func NewActor() {

}

func (a *Actor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetModes{})
		ctx.Send(ctx.Self(), &MsgGetRoutes{})
		ctx.Send(ctx.Self(), &MsgGetItinerary{})
	case *MsgGetModes:
		resp, err := utils.Get(a.httpClient, a.urlMode, a.userHttp, a.passHttp)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		var result []*Mode
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
	case *MsgGetRoutes:
		resp, err := utils.Get(a.httpClient, a.urlRoute, a.userHttp, a.passHttp)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		var result []*Route
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
	case *MsgGetItinerary:
		resp, err := utils.Get(a.httpClient, a.urlItinerary, a.userHttp, a.passHttp)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		var result []*Itinerary
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
		a.itineraryMap = make(ItineraryMap)
		for _, v := range result {
			a.itineraryMap[v.PaymentMediumCode] = v
		}
	case *MsgGetMap:
		if ctx.Sender() != nil {
			ctx.Respond(&MsgMap{Data: a.itineraryMap})
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	t2 := time.After(3 * time.Second)
	for {
		select {
		case <-t2:
			rootctx.Send(self, &MsgTick{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
