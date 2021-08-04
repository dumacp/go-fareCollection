package parameters

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

const (
	defaultURL         = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/dev-summary"
	defaultUsername    = "dev.nebulae"
	filterHttpQuery    = "?page=%d&count=%d&active=true"
	defaultPassword    = "uno.2.tres"
	dbpath             = "/SD/boltdb/parametersdb"
	databaseName       = "parametersdb"
	collectionNameData = "parameters"
)

type Actor struct {
	quit               chan int
	httpClient         *http.Client
	userHttp           string
	passHttp           string
	url                string
	id                 string
	db                 *actor.PID
	evs                *eventstream.EventStream
	parameters         *Parameters
	platformParameters *PlatformParameters
}

func NewActor(id string) actor.Actor {
	return &Actor{id: id}
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
		case *MsgParameters:
			return true
		}
		return false
	})
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in paramActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:

		//TODO: how get this params?
		a.url = defaultURL
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

		if a.parameters != nil && ctx.Sender() != nil {
			ctx.Respond(&MsgParameters{Data: a.parameters})
		}
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetParameters{})
	case *MsgGetParameters:
		isUpdateMap := false
		if err := func() error {
			url := fmt.Sprintf("%s/%s", a.url, a.id)
			resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
			if err != nil {
				return err
			}
			logs.LogBuild.Printf("Get url: %s", url)
			logs.LogBuild.Printf("Get response: %s", resp)
			var result PlatformParameters
			if err := json.Unmarshal(resp, &result); err != nil {
				return err
			}

			if a.platformParameters == nil || a.platformParameters.Timestamp < result.Timestamp {

				a.platformParameters = &result
				if a.parameters == nil {
					isUpdateMap = true
					a.parameters = new(Parameters)
				} else {
					if a.parameters.Timestamp < result.Timestamp {
						isUpdateMap = true
					}
				}
				a.parameters.FromPlatform(&result)

				data, err := json.Marshal(a.parameters)
				if err != nil {
					isUpdateMap = false
					return err
				}
				if a.db != nil && isUpdateMap {
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionNameData,
						ID:         a.parameters.ID,
						Data:       data,
					})
				}
			}
			return nil

		}(); err != nil {
			logs.LogError.Println(err)
		}
		if isUpdateMap {
			logs.LogBuild.Printf("params: %+v", a.parameters)
			ctx.Send(ctx.Self(), &MsgPublish{})
		}
	case *MsgPublish:
		if a.evs != nil {
			if a.parameters != nil {
				a.evs.Publish(&MsgParameters{Data: a.parameters})
			}
		}
	case *MsgGetInDB:
		if a.db == nil {
			break
		}
		ctx.Request(a.db, &database.MsgQueryData{
			Database:   databaseName,
			Collection: collectionNameData,
			PrefixID:   a.id,
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
			param := new(Parameters)
			if err := json.Unmarshal(msg.Data, param); err != nil {
				logs.LogError.Println(err)
				break
			}
			if a.parameters == nil || a.parameters.Timestamp < param.Timestamp {
				a.parameters = param
			}
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
			rootctx.Send(self, &MsgPublish{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
