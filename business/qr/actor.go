package qr

import (
	"math/rand"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

type Actor struct {
	ch  chan int
	ctx actor.Context
}

func NewActor() actor.Actor {

	a := &Actor{}
	a.ch = make(chan int, 0)
	return a
}

func (a *Actor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {

	case *actor.Started:
	case *MsgNewCodeQR:
		ctx.Send(ctx.Sender(), &MsgResponseCodeQR{Value: int(NewCode())})
	}
}

func (a *Actor) TickQR(ch <-chan int) {

	tick1 := time.NewTicker(20 * time.Second)
	defer tick1.Stop()

	go func() {
		for {
			select {
			case <-ch:
				return
			case <-tick1.C:
				a.ctx.Send(a.ctx.Parent(), &MsgResponseCodeQR{Value: int(NewCode())})
			}
		}
	}()
}

func NewCode() int32 {

	rand.Seed(time.Now().UnixNano())
	v1 := 12000 + rand.Int31n(10000)
	rand.Seed(time.Now().UnixNano())
	v2 := rand.Int31n(10000)

	return v1 + v2
}
