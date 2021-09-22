package fare

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/internal/utils"
)

func TestNewActor(t *testing.T) {

	utils.Url = "https://sibus.nebulae.com.co"
	sys := actor.NewActorSystem()

	rootctx := sys.Root

	props := actor.PropsFromProducer(NewActor)
	pid, err := rootctx.SpawnNamed(props, "fare-actor")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)

	type args struct {
		pid      *actor.PID
		rootctx  *actor.RootContext
		messages []interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				rootctx: rootctx,
				pid:     pid,
				messages: []interface{}{&MsgGetFare{
					LastFarePolicies: nil,
					ProfileID:        1,
					ItineraryID:      0,
					ModeID:           1,
					RouteID:          77,
					FromItineraryID:  0,
				}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, msg := range tt.args.messages {
				res, err := rootctx.RequestFuture(tt.args.pid, msg, time.Millisecond*900).Result()
				if err != nil {
					t.Log(err)
					continue
				}
				t.Logf("result: %#v", res)

			}
			time.Sleep(80 * time.Second)
		})
	}
}
