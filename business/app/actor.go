package app

import (
	"errors"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
)

type Actor struct {
	lastTag        uint64
	lastTimeDetect time.Time
	errorWriteTag  uint64
	actualTag      uint64
}

func NewActor() actor.Actor {
	return &Actor{}
}

func (a *Actor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *MsgTagDetected:
		if err := func() error {
			if a.actualTag == a.lastTag {
				if a.lastTimeDetect.Before(time.Now().Add(-5 * time.Second)) {
					return nil
				}
			}
			//read card
			defer ctx.Send(nil, nil)
			a.lastTag = msg.UID
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgTagRead:
		if err := func() error {
			if a.actualTag == a.errorWriteTag {
				//Commit Tag
				return nil
			}
			if err := ValidationTag(msg.Data); err != nil {
				if errors.Is(err, BalanceError) {
					//Send Msg Error Balance
					defer ctx.Send(nil, nil)
				}
				time.Sleep(3 * time.Second)
				return err
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgQRRead:
		if err := func() error {
			if a.actualTag == a.errorWriteTag {
				//Commit Tag
				return nil
			}
			if err := ValidationQr(msg.Data); err != nil {
				if errors.Is(err, QrError) {
					//Send Msg Error Balance
					defer ctx.Send(nil, nil)
				}
				time.Sleep(3 * time.Second)
				return err
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgTagWriteError:
		if err := func() error {
			//Send Msg Write error
			defer ctx.Send(nil, nil)
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	}
}
