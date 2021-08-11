package usostransporte

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
}

func NewActor(id string) actor.Actor {
	return &Actor{id: id}
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
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetParameters{})

	case *MsgUso:
		if err := func() error {
			uso := msg.Data
			data, err := json.Marshal(uso)
			if err != nil {
				return fmt.Errorf("send uso err: %w", err)
			}
			if resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, data); err != nil {
				if a.db != nil {
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionNameData,
						ID:         uso.ID,
						Data:       data,
					})
				} else {
					return fmt.Errorf("send uso (db is empty) err post: %s, %w", resp, err)
				}
				return fmt.Errorf("send uso err post: %s, %w", resp, err)
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

				resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, msg.Data)
				if err != nil {
					return fmt.Errorf("send uso err: %s, %w", resp, err)
				}
				ctx.Send(a.db, &database.MsgDeleteData{
					ID:         msg.ID,
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
