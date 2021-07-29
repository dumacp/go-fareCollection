package parameters

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/utils"
	"github.com/dumacp/go-logs/pkg/logs"
)

type Actor struct {
	quit       chan int
	httpClient *http.Client
	userHttp   string
	passHttp   string
	url        string
}

func NewActor(url, user, pass string) {

}

func (a *Actor) Receive(ctx actor.Context) {
	switch ctx.Message().(type) {
	case *actor.Started:
		a.quit = make(chan int)
		go tick(ctx, 60*time.Minute, a.quit)

	case *actor.Stopping:
		close(a.quit)
	case *MsgTick:
		ctx.Send(ctx.Self(), &MsgGetParameters{})
	case *MsgGetParameters:
		resp, err := utils.Get(a.httpClient, a.url, a.userHttp, a.passHttp, nil)
		if err != nil {
			logs.LogError.Println(err)
			break
		}
		var result map[string]interface{}
		if err := json.Unmarshal(resp, &result); err != nil {
			logs.LogError.Println(err)
			break
		}
	}
}

func tick(ctx actor.Context, timeout time.Duration, quit <-chan int) {
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTicker(timeout)
	for {
		select {
		case <-t1.C:
			rootctx.Send(self, &MsgTick{})
		case <-quit:
			return
		}
	}
}
