package usostransporte

import (
	"encoding/json"
	"errors"
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
	dbbkpath            = "/SD/boltdb/usosdb-BK"
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
	db   *actor.PID
	dbbk *actor.PID
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
			logstrans.LogError.Printf("open database usos err: %s", err)
			time.Sleep(6 * time.Second)
			panic(err)
		}
		if db != nil {
			a.db = db.PID()
		}
		dbbk, err := database.Open(ctx.ActorSystem().Root, dbbkpath)
		if err != nil {
			logstrans.LogError.Printf("open database usos-bk err: %s", err)
			time.Sleep(6 * time.Second)
			panic(err)
		}
		if dbbk != nil {
			a.dbbk = dbbk.PID()
		}

		a.quit = make(chan int)
		go tick(ctx, 30*time.Second, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetInDB{})
	case *MsgUso:
		if err := func() error {
			if a.db == nil {
				return fmt.Errorf("not database usos: %w", ErrorDB)
			}
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
			ctx.Request(a.dbbk, &database.MsgUpdateData{
				Database:   databaseName,
				Collection: collectionUsosData,
				ID:         uso.ID,
				Data:       data,
			})
			if resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, data); err != nil {
				ctx.Request(a.db, &database.MsgUpdateData{
					Database:   databaseName,
					Collection: collectionUsosData,
					ID:         uso.ID,
					Data:       data,
				})
				return err
			} else {
				logstrans.LogBuild.Printf("response send uso: %s", resp)
			}
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("send err: transaction uso with ID %s (payment ID = %s), err: %s",
				msg.Data.ID, msg.Data.PaymentMediumId, err)
			// log.Printf("data uso: %v", msg.Data)
			if errors.Is(err, ErrorDB) {
				ctx.Send(ctx.Parent(), &MsgErrorDB{
					Error: err.Error(),
				})
			}
		}
	case *MsgLock:
		if err := func() error {
			if a.db == nil {
				return fmt.Errorf("not database usos: %w", ErrorDB)
			}
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
			ctx.Request(a.dbbk, &database.MsgUpdateData{
				Database:   databaseName,
				Collection: collectionUsosData,
				ID:         uso.ID,
				Data:       data,
			})
			if resp, err := utils.Post(a.httpClient, a.urlLock, a.userHttp, a.passHttp, data); err != nil {
				ctx.Request(a.db, &database.MsgUpdateData{
					Database:   databaseName,
					Collection: collectionLocksData,
					ID:         uso.ID,
					Data:       data,
				})
				return err
			} else {
				logstrans.LogBuild.Printf("response send lock: %s", resp)
			}
			return nil
		}(); err != nil {
			logstrans.LogError.Printf("send err: transaction lock with ID %s (payment ID = %s), err: %s",
				msg.Data.ID, msg.Data.PaymentMediumId, err)
			// log.Printf("data lock: %v", msg.Data)
			if errors.Is(err, ErrorDB) {
				ctx.Send(ctx.Parent(), &MsgErrorDB{
					Error: err.Error(),
				})
			}
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
			if len(msg.Data) <= 0 {
				return nil
			}
			switch msg.Collection {
			case collectionUsosData:
				resp, err := utils.Post(a.httpClient, a.url, a.userHttp, a.passHttp, msg.Data)
				if err != nil {
					return err
				} else {
					logstrans.LogBuild.Printf("response send lock: %s", resp)
				}
			case collectionLocksData:
				resp, err := utils.Post(a.httpClient, a.urlLock, a.userHttp, a.passHttp, msg.Data)
				if err != nil {
					return err
				} else {
					logstrans.LogBuild.Printf("response send lock: %s", resp)
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
			// log.Printf("data uso: %s", msg.Data)
		}
	case *database.MsgNoAckPersistData:
		logstrans.LogError.Printf("error in DB, err: %s",
			msg.Error)
		ctx.Send(ctx.Parent(), &MsgErrorDB{
			Error: msg.Error,
		})
	case *MsgVerifyDB:
		if a.db != nil {
			res, err := ctx.RequestFuture(a.db, &database.MsgUpdateData{
				Database:   databaseName,
				Collection: collectionUsosData,
				ID:         "test",
				Data:       []byte("test data"),
			}, 1*time.Second).Result()
			if err != nil {
				logstrans.LogError.Printf("error in DB, err: %s",
					err)
				break
			}
			switch res.(type) {
			case *database.MsgAckPersistData:
				if ctx.Sender() != nil {
					ctx.Respond(&MsgOkDB{})
				}
			}
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	t2 := time.After(300 * time.Millisecond)
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
