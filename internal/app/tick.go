package app

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	t2 := time.After(3 * time.Second)
	// tqr := time.NewTicker(30 * time.Second)
	for {
		select {
		// case <-tqr.C:
		// 	rootctx.Send(self, &MsgNewRand{Value: int(NewCode())})
		case <-t2:
			rootctx.Send(self, &MsgTick{})
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
