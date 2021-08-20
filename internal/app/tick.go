package app

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	// t1 := time.NewTicker(timeout)
	// defer t1.Stop()
	// t2 := time.After(3 * time.Second)
	tWrite := time.NewTicker(300 * time.Second)
	defer tWrite.Stop()
	// tqr := time.NewTicker(30 * time.Second)
	for {
		select {
		// case <-tqr.C:
		// 	rootctx.Send(self, &MsgNewRand{Value: int(NewCode())})
		// case <-t2:
		// 	rootctx.Send(self, &MsgTick{})
		// case <-t1.C:
		// 	rootctx.Send(self, &MsgTick{})
		case <-tWrite.C:
			rootctx.Send(self, &MsgWriteAppParamas{})
		case <-quit:
			return
		}
	}
}
