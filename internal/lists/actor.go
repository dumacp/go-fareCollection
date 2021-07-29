package lists

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/database"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
)

type Actor struct {
	quit       chan int
	httpClient *http.Client
	userHttp   string
	passHttp   string
	url        string
	watchLists []string
	listMap    map[string]*List
	listInfo   map[string]*ListElement
	db         *actor.PID
}

const (
	defaultListURL     = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/payment-medium-lists"
	defaultUsername    = "dev.nebulae"
	defaultPassword    = "uno.2.tres"
	dbpath             = "/SD/boltdb/listdb"
	databaseName       = "listdb"
	collectionNameInfo = "listInfo"
	collectionNameData = "listData"
)

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in listActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:
		a.url = defaultListURL
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
		//TODO:
		// get http parameters
	case *actor.Stopping:
		close(a.quit)

	case *MsgGetListsInDB:
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
			list := new(List)
			if err := json.Unmarshal(msg.Data, list); err != nil {
				logs.LogError.Println(err)
				break
			}
			if a.listMap == nil {
				a.listMap = make(map[string]*List)
			}
			if v, ok := a.listMap[list.Code]; ok {
				if v.Metadata.UpdatedAt >= list.Metadata.UpdatedAt {
					break
				}
			}
			Populate(list)
			a.listMap[list.Code] = list
		}
	case *MsgTick:
		// ctx.Send(ctx.Self(), &MsgGetLists{})
		if len(a.watchLists) > 0 {
			for _, listid := range a.watchLists {
				ctx.Send(ctx.Self(), &MsgGetListById{ID: listid})
			}
		}
	case *MsgWatchList:
		if a.watchLists == nil {
			a.watchLists = make([]string, 0)
		}
		a.watchLists = append(a.watchLists, msg.ID)
		ctx.Send(ctx.Self(), &MsgTick{})
	case *MsgVerifyInList:

		func() {
			response := &MsgVerifyInListResponse{
				ListID: msg.ListID,
				ID:     nil,
			}
			defer func() {
				if ctx.Sender() != nil {
					ctx.Respond(response)
				}
			}()
			if a.listMap == nil {
				ctx.Send(ctx.Self(), &MsgGetListById{ID: msg.ListID})
				return
			}
			if _, ok := a.listMap[msg.ListID]; !ok {
				ctx.Send(ctx.Self(), &MsgGetListById{ID: msg.ListID})
				return
			}

			for _, id := range msg.ID {
				if list, ok := a.listMap[msg.ListID]; ok {
					if !list.Active {
						continue
					}
					if list.DataIds == nil {
						continue
					}
					if ok := list.DataIds.Search(id); ok {
						if response.ID == nil {
							response.ID = make([]int64, 0)
						}
						response.ID = append(response.ID, id)
					}
				}
			}
		}()
	case *MsgGetLists:
		resp, err := utils.Get(a.httpClient, a.url, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		logs.LogBuild.Printf("Get response: %s", resp)
		var result []*ListElement
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}

		a.listInfo = make(map[string]*ListElement)
		for _, v := range result {
			a.listInfo[v.Code] = v
		}
	case *MsgGetListById:
		url := fmt.Sprintf("%s/%s", a.url, msg.ID)
		resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		logs.LogBuild.Printf("Get response: %s", resp)
		list := new(List)
		if err := json.Unmarshal(resp, list); err != nil {
			logs.LogError.Println(err)
			break
		}
		if a.listMap == nil {
			a.listMap = make(map[string]*List)
		}
		if v, ok := a.listMap[list.Code]; ok {
			if v.Metadata.UpdatedAt >= list.Metadata.UpdatedAt {
				break
			}
		}
		if a.db != nil {
			ctx.Send(a.db, &database.MsgUpdateData{
				Database:   databaseName,
				Collection: collectionNameData,
				ID:         list.ID,
				Data:       resp,
			})
		}
		Populate(list)
		a.listMap[list.Code] = list

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
			rootctx.Send(self, &MsgGetListsInDB{})
		case <-t2:
			rootctx.Send(self, &MsgTick{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
