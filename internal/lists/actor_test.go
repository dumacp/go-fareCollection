package lists

import (
	"testing"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func TestNewActor(t *testing.T) {

	sys := actor.NewActorSystem()

	rootctx := sys.Root

	props := actor.PropsFromProducer(NewActor)
	pid, err := rootctx.SpawnNamed(props, "list-actor")
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
				messages: []interface{}{
					&MsgGetLists{},
					&MsgGetListById{ID: "6b7c067b-8f58-45f1-b70c-a1cd402c26e5"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, msg := range tt.args.messages {
				rootctx.Send(tt.args.pid, msg)
			}
			time.Sleep(10 * time.Second)
		})
	}
}
