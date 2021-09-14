package usostransporte

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/database"
	"github.com/dumacp/go-fareCollection/internal/logstrans"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
)

const (
	// transURL      = "https://fleet.nebulae.com.co/api/external-system-gateway/rest/fare-transaction"
	transURL        = "%s/api/external-system-gateway/rest/payment-medium-transaction"
	lockUrl         = "%s/api/external-system-gateway/rest/payment-medium-blocked"
	defaultUsername = "dev.nebulae"
	// filterHttpQuery    = "?page=%d&count=%d&active=true"
	defaultPassword     = "uno.2.tres"
	dbpath              = "/SD/boltdb/usosdb"
	databaseName        = "usosdb"
	collectionUsosData  = "usos"
	collectionLocksData = "locks"
)

type Actor struct {
	quit       chan int
	httpClient *http.Client
	userHttp   string
	passHttp   string
	url        string
	urlLock    string
	// id         string
	db *actor.PID
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in usosActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:

		//TODO: how get this params?
		a.url = fmt.Sprintf(transURL, utils.Url)
		a.urlLock = fmt.Sprintf(lockUrl, utils.Url)
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
		go tick(ctx, 60*time.Second, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetInDB{})
	case *MsgUso:
		if err := func() error {
			uso := msg.Data
			data, err := json.Marshal(uso)
			if err != nil {
				return fmt.Errorf("parse send uso err: %w", err)
			}
			if msg.Data.Error == nil || msg.Data.Error.Code > 0 {
				logstrans.LogTransWarn.Printf("uso transporte canceled: %s", data)
			} else {
				logstrans.LogTransInfo.Printf("uso transporte: %s", data)
			}
			if resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, data); err != nil {
				if a.db != nil {
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionUsosData,
						ID:         uso.ID,
						Data:       data,
					})
				} else {
					return fmt.Errorf("(db is empty) err post: %s, %w", resp, err)
				}
				return fmt.Errorf("err post: %s, %w", resp, err)
			} else {
				logstrans.LogBuild.Printf("response send uso: %s", resp)
			}
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("send err: transaction with ID %s (payment ID = %s), err: %s",
				msg.Data.ID, msg.Data.PaymentMediumId, err)
		}
	case *MsgLock:
		if err := func() error {
			uso := msg.Data
			data, err := json.Marshal(uso)
			if err != nil {
				return fmt.Errorf("parse send lock err: %w", err)
			}
			if msg.Data.Error == nil || msg.Data.Error.Code > 0 {
				logstrans.LogTransWarn.Printf("lock canceled: %s", data)
			} else {
				logstrans.LogTransWarn.Printf("lock: %s", data)
			}
			if resp, err := utils.Post(a.httpClient, a.urlLock, a.userHttp, a.passHttp, data); err != nil {
				if a.db != nil {
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionLocksData,
						ID:         uso.ID,
						Data:       data,
					})
				} else {
					return fmt.Errorf("(db is empty) err post: %s, %w", resp, err)
				}
				return fmt.Errorf("err post: %s, %w", resp, err)
			} else {
				logstrans.LogBuild.Printf("response send lock: %s", resp)
			}
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("send err: transaction with ID %s (payment ID = %s), err: %s",
				msg.Data.ID, msg.Data.PaymentMediumId, err)
		}
	case *MsgGetInDB:
		if a.db == nil {
			break
		}
		ctx.Request(a.db, &database.MsgQueryData{
			Database:   databaseName,
			Collection: collectionUsosData,
			PrefixID:   "",
			Reverse:    false,
		})
		ctx.Request(a.db, &database.MsgQueryData{
			Database:   databaseName,
			Collection: collectionLocksData,
			PrefixID:   "",
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
			case collectionUsosData:
				resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, msg.Data)
				if err != nil {
					return fmt.Errorf("err post: %s, %w", resp, err)
				}
			case collectionLocksData:
				resp, err := utils.Post(a.httpClient, a.urlLock, a.userHttp, a.passHttp, msg.Data)
				if err != nil {
					return fmt.Errorf("err post: %s, %w", resp, err)
				}
			}
			ctx.Send(a.db, &database.MsgDeleteData{
				ID:         msg.ID,
				Database:   databaseName,
				Collection: msg.Collection,
			})
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("send err: transactionID with ID %s, err: %s", msg.ID, err)
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
