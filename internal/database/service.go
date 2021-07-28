package database

import (
	"errors"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
)

type svc struct {
	db DB
}

type DBservice interface {
	Insert(id string, data []byte, database string, collection string) (string, error)
	Update(id string, data []byte, database string, collection string) (string, error)
	Get(id string, database string, collection string) ([]byte, error)
	Delete(id string, database string, collection string) error
	Query(database, collection, prefixID string, reverse bool, query func(data []byte) bool) error
	// PID() *actor.PID
}

func NewService(db DB) DBservice {
	service := &svc{db: db}

	return service
}

// func (db *svc) PID() *actor.PID {
// 	return db.ctx.Self()
// }

func (db *svc) Insert(id string, data []byte, database string, collection string) (string, error) {

	response, err := db.db.RootContext().RequestFuture(db.db.PID(), &MsgInsertData{
		ID:         id,
		Data:       data,
		Database:   database,
		Collection: collection,
	}, 1*time.Second).Result()
	if err != nil {

		return "", err
	}
	switch msg := response.(type) {
	case *MsgAckPersistData:
		return msg.ID, nil
	case *MsgNoAckPersistData:
		return "", errors.New(msg.Error)
	}
	return "", errors.New("whitout response")
}

func (db *svc) Update(id string, data []byte, database string, collection string) (string, error) {

	response, err := db.db.RootContext().RequestFuture(db.db.PID(), &MsgUpdateData{
		ID:         id,
		Data:       data,
		Database:   database,
		Collection: collection,
	}, 1*time.Second).Result()
	if err != nil {
		return "", err
	}
	switch msg := response.(type) {
	case *MsgAckPersistData:
		return msg.ID, nil
	case *MsgNoAckPersistData:
		return "", errors.New(msg.Error)
	}
	return "", errors.New("whitout response")
}

func (db *svc) Delete(id, database, collection string) error {

	response, err := db.db.RootContext().RequestFuture(db.db.PID(), &MsgDeleteData{
		ID:         id,
		Database:   database,
		Collection: collection,
	}, 1*time.Second).Result()
	if err != nil {
		return err
	}
	switch msg := response.(type) {
	case *MsgAckPersistData:
		return nil
	case *MsgNoAckPersistData:
		return errors.New(msg.Error)
	}
	return errors.New("whitout response")
}

func (db *svc) Get(id string, database string, collection string) ([]byte, error) {

	response, err := db.db.RootContext().RequestFuture(db.db.PID(), &MsgGetData{
		ID:         id,
		Database:   database,
		Collection: collection,
	}, 1*time.Second).Result()
	if err != nil {
		return nil, err
	}
	switch msg := response.(type) {
	case *MsgAckGetData:
		return msg.Data, nil
	case *MsgNoAckGetData:
		return nil, errors.New(msg.Error)
	}
	return nil, errors.New("whitout response")
}

func (db *svc) Query(database, collection, prefixID string, reverse bool, query func(data []byte) bool) error {

	type start struct{}
	timeout := 10 * time.Second
	sender := &actor.PID{}
	// var errFinal error
	props := actor.PropsFromFunc(func(ctx actor.Context) {
		logs.LogBuild.Printf("Message arrive in CHILD datab: %s, %T", ctx.Message(), ctx.Message())
		switch msg := ctx.Message().(type) {
		case *start:
			sender = ctx.Sender()
			ctx.Request(db.db.PID(), &MsgQueryData{
				PrefixID:   prefixID,
				Database:   database,
				Collection: collection,
				Reverse:    reverse,
			})
		case *MsgQueryResponse:
			if query(msg.Data) {
				ctx.Respond(&MsgQueryNext{})
				break
			}
			ctx.Send(sender, nil)
		case *MsgAckGetData:
			ctx.Send(sender, nil)
		case *MsgNoAckGetData:
			// errFinal = errors.New(msg.Error)
			ctx.Send(sender, errors.New(msg.Error))
		case *actor.Stopping:
		case *actor.Stopped:
			// ctx.Send(sender, nil)
		}
	})
	pid := db.db.RootContext().Spawn(props)
	defer func() {
		go db.db.RootContext().PoisonFuture(pid)
	}()

	req := db.db.RootContext().RequestFuture(pid, &start{}, timeout)
	res, err := req.Result()
	if err != nil {
		return err
	}
	switch msg := res.(type) {
	case error:
		return msg
	}
	return nil
}
