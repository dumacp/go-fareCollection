package se

import (
	"fmt"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/dumacp/smartcard"
	"github.com/google/uuid"
	"github.com/looplab/fsm"
)

type samActor struct {
	pid      *actor.PID
	rootctx  *actor.RootContext
	fm       *fsm.FSM
	behavior actor.Behavior
	reader   smartcard.IReader
	sam      SE
	ctx      actor.Context
}

type Actor interface {
	PID() *actor.PID
	RootContext() *actor.RootContext
}

func ActorSAM(ctx *actor.RootContext, r smartcard.IReader) (Actor, error) {
	app := &samActor{}
	app.reader = r

	app.initFSM()
	props := actor.PropsFromFunc(app.Receive)

	if ctx == nil {
		ctx = actor.NewActorSystem().Root
	}
	app.rootctx = ctx
	uid, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	pid, err := ctx.SpawnNamed(props, fmt.Sprintf("sam-actor-%s", uid.String()))
	if err != nil {
		return nil, err
	}
	app.pid = pid
	app.behavior = make(actor.Behavior, 0)
	app.behavior.Become(app.CloseState)
	app.behavior = actor.NewBehavior()
	app.behavior.Become(app.CloseState)
	app.initFSM()
	return app, nil
}

func (app *samActor) PID() *actor.PID {
	return app.pid
}
func (app *samActor) RootContext() *actor.RootContext {
	return app.rootctx
}

func (a *samActor) Receive(ctx actor.Context) {
	// logs.LogBuild.Printf("Message arrived in appActor: %s, %T", ctx.Message(), ctx.Message())
	a.ctx = ctx
	a.behavior.Receive(ctx)
}

func (a *samActor) CloseState(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in samActor, behavior (CloseState): %s, %T", ctx.Message(), ctx.Message())
	switch ctx.Message().(type) {
	case *actor.Started:
		ctx.Send(ctx.Self(), &MsgOpen{})
	case *MsgOpen:
		if err := func() error {
			c, err := NewSamAV2(a.reader)
			if err != nil {
				return err
			}
			if err := c.Connect(); err != nil {
				return err
			}
			a.sam = c
			a.fm.Event(eOpenCmd)
			logs.LogBuild.Printf("sam UID: [% X]", c.Serial())
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			a.fm.Event(eError, err)
		}
	}
}

func (a *samActor) WaitState(ctx actor.Context) {
	logs.LogBuild.Printf("Message arrived in samActor (%s), behavior (WaitState): %+v, %T, %s",
		ctx.Self().GetId(), ctx.Message(), ctx.Message(), ctx.Sender())
	switch msg := ctx.Message().(type) {
	case *MsgClose:
		if err := func() error {
			if err := a.sam.Disconnect(); err != nil {
				return err
			}
			a.fm.Event(eClosed)
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
		}
	case *MsgEncryptRequest:
		if err := func() error {
			cipher, err := a.sam.Encrypt(msg.Data, msg.IV, msg.DevInput, msg.KeySlot)
			if err != nil {
				return err
			}
			if ctx.Sender() != nil {
				ctx.Respond(&MsgEncryptResponse{
					Cipher: cipher,
				})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			if ctx.Sender() != nil {
				time.Sleep(3 * time.Second)
				ctx.Respond(&MsgAck{Error: err.Error()})
			}
			a.fm.Event(eError, err)
		}
	case *MsgDecryptRequest:
		if err := func() error {
			cipher, err := a.sam.Decrypt(msg.Data, msg.IV, msg.DevInput, msg.KeySlot)
			if err != nil {
				return err
			}
			if ctx.Sender() != nil {
				ctx.Respond(&MsgDecryptResponse{
					Plain: cipher,
				})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			if ctx.Sender() != nil {
				time.Sleep(3 * time.Second)
				ctx.Respond(&MsgAck{Error: err.Error()})
			}
			a.fm.Event(eError, err)
		}
	case *MsgDumpSecretKeyRequest:
		if err := func() error {
			key, err := a.sam.DumpSecretKey(msg.KeySlot)
			if err != nil {
				return err
			}
			if ctx.Sender() != nil {
				ctx.Respond(&MsgDumpSecretKeyResponse{
					Data: key,
				})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			if ctx.Sender() != nil {
				time.Sleep(3 * time.Second)
				ctx.Respond(&MsgAck{Error: err.Error()})
			}
			a.fm.Event(eError, err)
		}
	case *MsgCreateKeyRequest:
		if err := func() error {
			if err := a.sam.GenerateKey(msg.KeySlot, msg.Alg); err != nil {
				return err
			}
			logs.LogBuild.Printf("Key Create, %d", msg.KeySlot)
			if ctx.Sender() != nil {
				ctx.Respond(&MsgAck{})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			// if ctx.Sender() != nil {
			// 	time.Sleep(3 * time.Second)
			// 	ctx.Respond(&MsgAck{Error: err.Error()})
			// }
			a.fm.Event(eError, err)
		}
	case *MsgImportKeyRequest:
		if err := func() error {
			// if a.sam.EnableKeys() msg.KeySlot
			if err := a.sam.ImportKey(msg.Data, msg.KeySlot, msg.Alg); err != nil {
				return err
			}
			logs.LogBuild.Printf("Key import, %d", msg.KeySlot)
			if ctx.Sender() != nil {
				ctx.Respond(&MsgAck{})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			// if ctx.Sender() != nil {
			// 	time.Sleep(3 * time.Second)
			// 	ctx.Respond(&MsgAck{Error: err.Error()})
			// }
			a.fm.Event(eError, err)
		}
	case *MsgEnableKeysRequest:
		if err := func() error {
			keys, err := a.sam.EnableKeys()
			if err != nil {
				return err
			}
			logs.LogBuild.Printf("enable Keys, %+v", keys)
			if ctx.Sender() != nil {
				ctx.Request(ctx.Sender(), &MsgEnableKeysResponse{Data: keys})
			}
			return nil
		}(); err != nil {
			logs.LogError.Println(err)
			// if ctx.Sender() != nil {
			// 	time.Sleep(3 * time.Second)
			// 	ctx.Respond(&MsgAck{Error: err.Error()})
			// }
			a.fm.Event(eError, err)
		}
	}
}
