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
	filterHttpQuery    = "?page=%d&count=%d&queryTotalResultCount=%v"
	defaultUsername    = "dev.nebulae"
	defaultPassword    = "uno.2.tres"
	dbpath             = "/SD/boltdb/faredb"
	databaseName       = "faredb"
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
	logs.LogBuild.Printf("Message arrived in fareActor: %s, %T, %s",
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
		go func(ctx actor.Context) {
			if err := func() error {
				count := 30
				numPages := 1000
				// 			var paginationInput = `{
				// "paginationInput": {
				// 	"page": %d,
				// 	"count": %d,
				// 	"queryTotalResultCount": true
				// 	}
				// }`
				actives := make(map[int]int)
				for i := range make([]int, numPages) {
					// dataToPagination := fmt.Sprintf(paginationInput, count, i)
					filter := fmt.Sprintf(filterHttpQuery, count, i, true)
					url := fmt.Sprintf("%s%s", a.url, filter)
					resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp,
						nil)
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
					for _, fareNode := range result.Listing {

						actives[fareNode.ID] = 0
						if _, ok := a.farePolicies[fareNode.ID]; ok {
							continue
						}
						a.farePolicies[fareNode.ID] = fareNode
						if a.db != nil {
							v, err := json.Marshal(fareNode)
							if err != nil {
								continue
							}
							ctx.Send(a.db, &database.MsgUpdateData{
								Database:   databaseName,
								Collection: collectionNameData,
								ID:         fmt.Sprintf("%s:%d", fareNode.FarePolicyID, fareNode.ID),
								Data:       v,
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
				logs.LogInfo.Printf("fare map: %#v", a.fareMap)
				return nil
			}(); err != nil {
				logs.LogError.Println(err)
			}
		}(ctx)
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
				ProfileID:       msg.ProfileID,
				LastFare:        make([]*FareNode, 0),
			}
			if msg.FromItineraryID > 0 {
				if a.itineraryMap != nil &&
					a.itineraryMap[msg.ItineraryID] != nil {
					q.FromModeID = a.itineraryMap[msg.ItineraryID].ModePaymentMediumCode
					q.FromRouteID = a.itineraryMap[msg.ItineraryID].RoutePaymentMediumCode
				} else {
					logs.LogError.Printf("itinerary not found: %d", msg.FromItineraryID)
					//TODO: ?
					// if a.pidItinerary != nil {
					// 	ctx.Request(a.pidItinerary, &itinerary.MsgGetMap{})
					// }
				}
			}
			keySort := make([]int64, 0)
			for k := range msg.LastFarePolicies {
				keySort = append(keySort, k)
			}
			sort.Slice(keySort, func(i, j int) bool { return keySort[i] < keySort[j] })

			logs.LogBuild.Printf("fare map: %#v", a.fareMap)
			logs.LogBuild.Printf("fare policies: %#v", a.farePolicies)
			logs.LogBuild.Printf("last  fare policies: %#v", msg.LastFarePolicies)
			// foundPlain := false
			for _, k := range keySort {
				fare, ok := a.farePolicies[msg.LastFarePolicies[k]]
				if !ok {
					break
				}
				// if !foundPlain {
				if fare.Type == PLAIN {
					// foundPlain = true
					q.LastFare = append(q.LastFare, fare)
					q.LastTimePlain = time.Unix(int64(k), 0)
					break
				}
				// continue
				// }
				q.LastFare = append(q.LastFare, fare)
			}
			logs.LogBuild.Printf("fare query: %#v", q)
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
				ctx.Respond(&MsgFare{
					Fare:         fare.Fare,
					FarePolicyID: fare.ID,
					// ItineraryID:  fare.ItineraryID,
				})
			}
		}

		// calculate(msg.LastFarePolicies, a.farePolicies)
	case *MsgGetFareInDB:
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
			fare := new(FareNode)
			if err := json.Unmarshal(msg.Data, fare); err != nil {
				logs.LogError.Println(err)
				break
			}
			if a.itineraryMap == nil {
				a.fareMap = make(FareMap)
			}
			a.farePolicies[fare.ID] = fare
			a.fareMap = CreateTree(a.farePolicies)
		}
		logs.LogBuild.Printf("fare map: %#v", a.fareMap)
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
