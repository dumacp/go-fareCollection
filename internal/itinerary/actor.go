package itinerary

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/database"
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
	db           *actor.PID
}

const (
	defaultModeURL      = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_modes"
	defaultRouteURL     = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_routes"
	defaultItineraryURL = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_itineraries"
	defaultUsername     = "dev.nebulae"
	defaultPassword     = "uno.2.tres"
	dbpath              = "/SD/boltdb/itinerarydb"
	databaseName        = "itinerarydb"
	collectionNameData  = "itineraries"
)

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in readerActor: %+v, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:

		a.urlMode = defaultModeURL
		a.urlRoute = defaultRouteURL
		a.urlItinerary = defaultItineraryURL
		a.passHttp = defaultPassword
		a.userHttp = defaultUsername

		db, err := database.Open(ctx.ActorSystem().Root, dbpath)
		if err != nil {
			logs.LogError.Println(err)
		}
		if db != nil {
			a.db = db.PID()
		}

		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetModes{})
		ctx.Send(ctx.Self(), &MsgGetRoutes{})
		ctx.Send(ctx.Self(), &MsgGetItinerary{})
	case *MsgGetModes:
		resp, err := utils.Get(a.httpClient, a.urlMode, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		logs.LogBuild.Printf("Get response: %s", resp)
		var result []*Mode
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
	case *MsgGetRoutes:
		resp, err := utils.Get(a.httpClient, a.urlRoute, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		logs.LogBuild.Printf("Get response: %s", resp)
		var result []*Route
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
	case *MsgGetItinerary:
		resp, err := utils.Get(a.httpClient, a.urlItinerary, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		logs.LogBuild.Printf("Get response: %s", resp)
		var result []*Itinerary
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
		actives := make(map[int]int)
		a.itineraryMap = make(ItineraryMap)
		for _, v := range result {
			actives[v.ModePaymentMediumCode] = 0
			a.itineraryMap[v.PaymentMediumCode] = v
			if iti, ok := a.itineraryMap[v.PaymentMediumCode]; ok {
				if iti.Metadata.UpdatedAt >= v.Metadata.UpdatedAt {
					continue
				}
			}
			if a.db != nil {
				ctx.Send(a.db, &database.MsgUpdateData{
					Database:   databaseName,
					Collection: collectionNameData,
					ID:         v.ID,
					Data:       resp,
				})
			}
		}

		if a.db != nil {
			for k, v := range a.itineraryMap {
				if _, ok := actives[k]; !ok {

					ctx.Send(a.db, &database.MsgDeleteData{
						ID:         v.ID,
						Database:   databaseName,
						Collection: collectionNameData,
					})
				}
			}
		}

	case *MsgGetMap:
		if ctx.Sender() != nil {
			ctx.Respond(&MsgMap{Data: a.itineraryMap})
		}
	case *MsgGetInDB:
		if a.db == nil {
			break
		}
		ctx.Request(a.db, &database.MsgQueryData{
			Database:   databaseName,
			Collection: collectionNameData,
			PrefixID:   "",
			Reverse:    false,
		})
	case *database.MsgQueryResponse:

		if ctx.Sender() != nil {
			ctx.Send(ctx.Sender(), &database.MsgQueryNext{})
		}

		if msg.Data == nil {
			break
		}
		switch msg.Collection {
		case collectionNameData:
			iti := new(Itinerary)
			if err := json.Unmarshal(msg.Data, iti); err != nil {
				logs.LogError.Println(err)
				break
			}
			if a.itineraryMap == nil {
				a.itineraryMap = make(ItineraryMap)
			}
			if v, ok := a.itineraryMap[iti.PaymentMediumCode]; ok {
				if v.Metadata.UpdatedAt >= iti.Metadata.UpdatedAt {
					break
				}
			}
			a.itineraryMap[iti.PaymentMediumCode] = iti
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	t2 := time.After(3 * time.Second)
	t3 := time.After(2 * time.Second)
	for {
		select {
		case <-t3:
			rootctx.Send(self, &MsgGetInDB{})
		case <-t2:
			rootctx.Send(self, &MsgTick{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
