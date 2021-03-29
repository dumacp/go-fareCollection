package database

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
	"github.com/etcd-io/bbolt"
)

type Actor struct {
	db *bbolt.DB
}

func NewActor(path string) actor.Actor {

	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {

	case *actor.Started:
		logs.LogBuild.Printf("actor %q started", ctx.Self().GetId())

	case MsgPersistData:
	}
}
