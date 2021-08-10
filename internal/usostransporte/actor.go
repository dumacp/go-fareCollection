package usostransporte

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/database"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	defaultURL         = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/dev-summary"
	defaultUsername    = "dev.nebulae"
	filterHttpQuery    = "?page=%d&count=%d&active=true"
	defaultPassword    = "uno.2.tres"
	dbpath             = "/SD/boltdb/usosdb"
	databaseName       = "usosdb"
	collectionNameData = "usos"
)

type Actor struct {
	quit       chan int
	httpClient *http.Client
	userHttp   string
	passHttp   string
	url        string
	id         string
	db         *actor.PID
	// evs        *eventstream.EventStream
}

func NewActor(id string) actor.Actor {
	return &Actor{id: id}
}

// func subscribe(ctx actor.Context, evs *eventstream.EventStream) {
// 	rootctx := ctx.ActorSystem().Root
// 	pid := ctx.Sender()
// 	self := ctx.Self()

// 	fn := func(evt interface{}) {
// 		rootctx.RequestWithCustomSender(pid, evt, self)
// 	}
// 	sub := evs.Subscribe(fn)
// 	sub.WithPredicate(func(evt interface{}) bool {
// 		switch evt.(type) {
// 		case nil:
// 			return true
// 		}
// 		return false
// 	})
// }

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
	// case *MsgSubscribe:
	// 	if a.evs == nil {
	// 		a.evs = eventstream.NewEventStream()
	// 	}
	// 	subscribe(ctx, a.evs)

	// 	if a.parameters != nil && ctx.Sender() != nil {
	// 		ctx.Respond(&MsgParameters{Data: a.parameters})
	// 	}
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetParameters{})

	// case *MsgPublish:
	// 	if a.evs != nil {
	// 		if a.parameters != nil {
	// 			a.evs.Publish(&MsgParameters{Data: a.parameters})
	// 		}
	// 	}
	case *MsgUso:
		if err := func() error {
			uso := msg.Data
			if err := send(uso); err != nil {
				if a.db != nil {
					data, err2 := json.Marshal(uso)
					if err != nil {
						return err2
					}
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionNameData,
						ID:         uso.ID,
						Data:       data,
					})
				}
				return err
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
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
		if err := func() error {
			if ctx.Sender() != nil {
				ctx.Send(ctx.Sender(), &database.MsgQueryNext{})
			}
			if msg.Data == nil {
				return nil
			}
			switch msg.Collection {
			case collectionNameData:
				uso := new(UsoTransporte)
				if err := json.Unmarshal(msg.Data, uso); err != nil {
					return err
				}
				if err := send(uso); err != nil {
					return err
				}
				ctx.Send(a.db, &database.MsgDeleteData{
					ID:         uso.ID,
					Database:   databaseName,
					Collection: collectionNameData,
				})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
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
