package itinerary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/eventstream"
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
	routeMap     map[int]*Route
	modeMap      map[int]*Mode
	db           *actor.PID
	evs          *eventstream.EventStream
}

const (
	defaultModeURL      = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_modes"
	defaultRouteURL     = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_routes"
	defaultItineraryURL = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device_itineraries"
	defaultUsername     = "dev.nebulae"
	filterHttpQuery     = "?page=%d&count=%d&active=true"
	defaultPassword     = "uno.2.tres"
	dbpath              = "/SD/boltdb/itinerarydb"
	databaseName        = "itinerarydb"
	collectionNameData  = "itineraries"
)

func NewActor() actor.Actor {
	return &Actor{}
}

func subscribe(ctx actor.Context, evs *eventstream.EventStream) {
	rootctx := ctx.ActorSystem().Root
	pid := ctx.Sender()
	self := ctx.Self()

	fn := func(evt interface{}) {
		rootctx.RequestWithCustomSender(pid, evt, self)
	}
	sub := evs.Subscribe(fn)
	sub.WithPredicate(func(evt interface{}) bool {
		switch evt.(type) {
		case *MsgMap:
			return true
		}
		return false
	})
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in itineraryActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:

		//TODO: how get this params?
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
	case *MsgSubscribe:
		if a.evs == nil {
			a.evs = eventstream.NewEventStream()
		}
		subscribe(ctx, a.evs)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetModes{})
		ctx.Send(ctx.Self(), &MsgGetRoutes{})
		ctx.Send(ctx.Self(), &MsgGetItinerary{})
	case *MsgGetModes:
		if err := func() error {
			count := 30
			numPages := 1000
			for i := range make([]int, numPages) {
				filter := fmt.Sprintf(filterHttpQuery, count, i)
				url := fmt.Sprintf("%s%s", a.urlMode, filter)
				resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get response: %s", resp)
				var result []*Mode
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if a.modeMap == nil {
					a.modeMap = make(map[int]*Mode)
				}
				for _, v := range result {
					a.modeMap[v.PaymentMediumCode] = v
				}

				if len(result) < count {
					break
				}
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgGetRoutes:
		if err := func() error {
			count := 30
			numPages := 1000
			for i := range make([]int, numPages) {
				filter := fmt.Sprintf(filterHttpQuery, count, i)
				url := fmt.Sprintf("%s%s", a.urlRoute, filter)
				resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get response: %s", resp)
				var result []*Route
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if a.routeMap == nil {
					a.routeMap = make(map[int]*Route)
				}
				for _, v := range result {
					a.routeMap[v.PaymentMediumCode] = v
				}
				if len(result) < count {
					break
				}
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgGetItinerary:
		isUpdateMap := false
		if err := func() error {
			count := 30
			numPages := 1000
			actives := make(map[int]int)
			for i := range make([]int, numPages) {
				filter := fmt.Sprintf(filterHttpQuery, count, i)
				url := fmt.Sprintf("%s%s", a.urlItinerary, filter)
				resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get response: %s", resp)
				var result []*Itinerary
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if a.itineraryMap == nil {
					a.itineraryMap = make(ItineraryMap)
				}
				for _, v := range result {
					actives[v.PaymentMediumCode] = 0
					if iti, ok := a.itineraryMap[v.PaymentMediumCode]; ok {
						if iti.Metadata.UpdatedAt >= v.Metadata.UpdatedAt {
							continue
						}
					}
					isUpdateMap = true
					a.itineraryMap[v.PaymentMediumCode] = v
					if a.db != nil {
						data, err := json.Marshal(v)
						if err != nil {
							continue
						}
						ctx.Send(a.db, &database.MsgUpdateData{
							Database:   databaseName,
							Collection: collectionNameData,
							ID:         v.ID,
							Data:       data,
						})
					}
				}
				if len(result) < count {
					break
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
			return nil

		}(); err != nil {
			logs.LogError.Println(err)
		}
		if isUpdateMap {
			ctx.Send(ctx.Self(), &MsgPublish{})
		}
	case *MsgPublish:
		if a.evs != nil {
			a.evs.Publish(&MsgMap{Data: a.itineraryMap})
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
