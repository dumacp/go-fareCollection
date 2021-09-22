package mediums

import (
	"encoding/json"
	"errors"
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
	paymentMediumURL             = "%s/api/external-system-gateway/rest/payment-medium-types"
	defaultUsername              = "dev.nebulae"
	filterHttpQuery              = "?page=%d&count=%d&active=true"
	defaultPassword              = "uno.2.tres"
	dbpath                       = "/SD/boltdb/mediumsdb"
	databaseName                 = "mediumsdb"
	collectionPaymentMediumsData = "Mediums"
)

type Actor struct {
	quit           chan int
	httpClient     *http.Client
	userHttp       string
	passHttp       string
	mediumUrl      string
	db             *actor.PID
	paymentMediums map[string]*PaymentMedium
}

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
		case *MsgPaymentMedium:
			return true
		}
		return false
	})
}

func (a *Actor) Receive(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in mediumActor: %s, %T, %s",
		ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *actor.Started:

		//TODO: how get this params?
		a.mediumUrl = fmt.Sprintf(paymentMediumURL, utils.Url)

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
		go tick(ctx, 48*time.Hour, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetPaymentMedium{})
	case *MsgGetPaymentMedium:
		go func(ctx actor.Context) {
			isUpdateMap := false
			if err := func() error {
				url := a.mediumUrl
				resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get url: %s", url)
				logs.LogBuild.Printf("Get response: %s", resp)
				var result []*PaymentMedium
				if err := json.Unmarshal(resp, &result); err != nil {
					return err
				}
				if len(result) <= 0 {
					return errors.New("error empty paymentMedims result")
				}

				if a.paymentMediums == nil {
					a.paymentMediums = make(map[string]*PaymentMedium)
				}
				for _, v := range result {
					if !v.Active {
						if _, ok := a.paymentMediums[v.Code]; ok {
							delete(a.paymentMediums, v.Code)
							if a.db != nil {
								ctx.Send(a.db, &database.MsgDeleteData{
									Database:   databaseName,
									Collection: collectionPaymentMediumsData,
									ID:         v.Code,
								})
							}
						}
						continue
					}
					if old_v, ok := a.paymentMediums[v.Code]; ok {
						if old_v.Metadata.UpdateAt >= v.Metadata.UpdateAt {
							continue
						}
					}
					// a.paymentMediums[v.Code] = v
					isUpdateMap = true
					ctx.Send(ctx.Self(), &MsgGetPaymentMediumByID{
						ID: v.ID,
					})
				}

				// data, err := json.Marshal(result)
				// if err != nil {
				// 	isUpdateMap = false
				// 	return err
				// }
				// if a.db != nil && isUpdateMap {
				// 	ctx.Send(a.db, &database.MsgUpdateData{
				// 		Database:   databaseName,
				// 		Collection: collectionPaymentMediumsData,
				// 		ID:         "list",
				// 		Data:       data,
				// 	})
				// }

				return nil

			}(); err != nil {
				logs.LogError.Println(err)
				go func() {
					if len(a.paymentMediums) <= 0 {
						time.Sleep(30 * time.Second)
						ctx.Send(ctx.Self(), &MsgGetPaymentMedium{})
					}
				}()
				return
			}
			if isUpdateMap {
				logs.LogBuild.Printf("payment medium: %+v", a.paymentMediums)
				// ctx.Send(ctx.Self(), &MsgPublish{})
			}
		}(ctx)
	case *MsgGetPaymentMediumByID:
		go func(ctx actor.Context) {
			isUpdateMap := false
			if err := func() error {
				url := fmt.Sprintf("%s/%s", a.mediumUrl, msg.ID)
				resp, err := utils.Get(a.httpClient, url, a.userHttp, a.passHttp, nil)
				if err != nil {
					return err
				}
				logs.LogBuild.Printf("Get url: %s", url)
				logs.LogBuild.Printf("Get response: %s", resp)
				result := new(PaymentMedium)
				if err := json.Unmarshal(resp, result); err != nil {
					return err
				}

				if a.paymentMediums == nil {
					isUpdateMap = true
					if result.Active {
						if a.paymentMediums == nil {
							a.paymentMediums = make(map[string]*PaymentMedium)
						}
						a.paymentMediums[result.Code] = result
					}

				} else {
					if !result.Active {
						if _, ok := a.paymentMediums[result.Code]; ok {
							delete(a.paymentMediums, result.Code)
							if a.db != nil {
								ctx.Send(a.db, &database.MsgDeleteData{
									Database:   databaseName,
									Collection: collectionPaymentMediumsData,
									ID:         result.Code,
								})
							}
						}
						return nil
					}
					if old_v, ok := a.paymentMediums[result.Code]; ok {
						if old_v.Metadata.UpdateAt >= result.Metadata.UpdateAt {
							return nil
						}
					}
					a.paymentMediums[result.Code] = result
					isUpdateMap = true
				}

				data, err := json.Marshal(result)
				if err != nil {
					isUpdateMap = false
					return err
				}
				if a.db != nil && isUpdateMap {
					ctx.Send(a.db, &database.MsgUpdateData{
						Database:   databaseName,
						Collection: collectionPaymentMediumsData,
						ID:         result.Code,
						Data:       data,
					})
				}

				return nil

			}(); err != nil {
				logs.LogError.Println(err)
				go func() {
					if a.paymentMediums == nil {
						time.Sleep(30 * time.Second)
						ctx.Send(ctx.Self(), msg)
					}
				}()
				return
			}
			if isUpdateMap {
				logs.LogBuild.Printf("payment medium: %+v", a.paymentMediums)
				ctx.Send(ctx.Self(), &MsgPublishPaymentMedium{})
			}
		}(ctx)
	case *MsgGetInDB:
		if a.db == nil {
			break
		}
		ctx.Request(a.db, &database.MsgQueryData{
			Database:   databaseName,
			Collection: collectionPaymentMediumsData,
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
		case collectionPaymentMediumsData:
			paym := new(PaymentMedium)
			if err := json.Unmarshal(msg.Data, paym); err != nil {
				logs.LogError.Println(err)
				break
			}
			if a.paymentMediums == nil {
				a.paymentMediums = make(map[string]*PaymentMedium)
			}
			a.paymentMediums[paym.Code] = paym
		}
	case *MsgRequestStatus:
		if ctx.Sender() != nil {
			break
		}
		if len(a.paymentMediums) > 0 {
			ctx.Respond(&MsgStatus{State: true})
		} else {
			ctx.Respond(&MsgStatus{State: false})
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	defer t1.Stop()
	t2 := time.After(200 * time.Millisecond)
	t3 := time.After(100 * time.Millisecond)
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
