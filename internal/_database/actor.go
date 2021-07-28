package database

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
	"go.etcd.io/bbolt"
)

type Actor struct {
	db *bbolt.DB
}

func NewActor(path string) actor.Actor {

	a := &Actor{}
	var err error
	a.db, err = OpenDB(path)
	if err != nil {
		panic(err)
	}
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {

	case *actor.Started:
		logs.LogBuild.Printf("actor %q started", ctx.Self().GetId())

	case MsgPersistData:
		if err := Put(a.db, msg.Database, msg.ID, msg.Indexes, &msg.TimeStamp, msg.Data); err != nil {
			if ctx.Sender() != nil {
				ctx.Respond(&MsgAckPersistData{ID: msg.ID, Error: err.Error(), Succes: false})
			}
		} else {
			if ctx.Sender() != nil {
				ctx.Respond(&MsgAckPersistData{ID: msg.ID, Error: "", Succes: true})
			}
		}
	}
}
