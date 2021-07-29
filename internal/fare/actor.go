package fare

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/database"
	"github.com/dumacp/go-fareCollection/internal/itinerary"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	defaultListURL     = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/device-fares"
	defaultUsername    = "dev.nebulae"
	defaultPassword    = "uno.2.tres"
	dbpath             = "/SD/boltdb/faredb"
	databaseName       = "listdb"
	collectionNameData = "farePolicies"
)

type Actor struct {
	farePolicies map[int]*FareNode
	fareMap      FareMap
	itineraryMap itinerary.ItineraryMap
	pidItinerary *actor.PID
	quit         chan int
	httpClient   *http.Client
	userHttp     string
	passHttp     string
	url          string
	db           *actor.PID
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in readerActor: %+v, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.farePolicies = make(map[int]*FareNode)

		a.url = defaultListURL
		a.passHttp = defaultPassword
		a.userHttp = defaultUsername

		db, err := database.Open(ctx.ActorSystem().Root, dbpath)
		if err != nil {
			logs.LogError.Println(err)
		}
		a.db = db.PID()

		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetFarePolicies{})
	case *MsgGetFarePolicies:
		//TODO:
		//Get Fare Policies from platform
		if err := func() error {
			count := 30
			numPages := 1000
			var paginationInput = `{
"paginationInput": {
	"page": %d,
	"count": %d,
	"queryTotalResultCount": true
	}
}`
			actives := make(map[int]int)
			for i := range make([]int, numPages) {
				dataToPagination := fmt.Sprintf(paginationInput, count, i)
				resp, err := utils.Get(a.httpClient, a.url, a.userHttp, a.passHttp,
					[]byte(dataToPagination))
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get response: %s", resp)
				var result = struct {
					Listing               []*FareNode `json:"listing"`
					QueryTotalResultCount int         `json:"queryTotalResultCount"`
				}{}

				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}

				if a.farePolicies == nil {
					a.farePolicies = make(map[int]*FareNode)
				}
				for _, v := range result.Listing {
					actives[v.ID] = 0
					if _, ok := a.farePolicies[v.ID]; ok {
						continue
					}
					a.farePolicies[v.ID] = v
					if a.db != nil {
						ctx.Send(a.db, &database.MsgUpdateData{
							Database:   databaseName,
							Collection: collectionNameData,
							ID:         v.FarePolicyID,
							Data:       resp,
						})
					}

				}
				if result.QueryTotalResultCount < count {
					break
				}
			}

			if a.db != nil {
				for k, v := range a.farePolicies {
					if _, ok := actives[k]; !ok {
						ctx.Send(a.db, &database.MsgDeleteData{
							ID:         v.FarePolicyID,
							Database:   databaseName,
							Collection: collectionNameData,
						})
					}
				}
			}

			a.fareMap = CreateTree(a.farePolicies)
			logs.LogBuild.Printf("fare map: %#v", a.fareMap)
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *itinerary.MsgMap:
		if ctx.Sender() != nil {
			a.pidItinerary = ctx.Sender()
		}
		a.itineraryMap = msg.Data
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
					q.FromModeID = a.itineraryMap[msg.ItineraryID].ModePaymentMediumCode
					q.FromRouteID = a.itineraryMap[msg.ItineraryID].RoutePaymentMediumCode
				}
			}
			keySort := make([]int64, 0)
			for k := range msg.LastFarePolicies {
				keySort = append(keySort, k)
			}
			sort.Slice(keySort, func(i, j int) bool { return keySort[i] < keySort[j] })

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

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	t2 := time.After(3 * time.Second)
	t3 := time.After(2 * time.Second)
	for {
		select {
		case <-t3:
			rootctx.Send(self, &MsgGetFareInDB{})
		case <-t2:
			rootctx.Send(self, &MsgTick{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
