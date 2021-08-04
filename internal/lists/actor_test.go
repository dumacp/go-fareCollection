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
					// &MsgGetLists{},
					// &MsgGetListById{ID: "6b7c067b-8f58-45f1-b70c-a1cd402c26e5"},
					// &MsgVerifyInList{ListID: "LIST001", ID: []int64{44913149217}},
					&MsgWatchList{ID: "PMLIST"},
					&MsgVerifyInList{ListID: "PMLIST", ID: []int64{44913149217, 423155168}},
					&MsgVerifyInList{ListID: "PMLIST", ID: []int64{44913149217, 423155168}},
					&MsgGetListById{ID: "904e7c85-08f4-47de-bd88-71458b9e2faf", Code: "PMLIST"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, msg := range tt.args.messages {
				res, err := rootctx.RequestFuture(tt.args.pid, msg, time.Millisecond*900).Result()
				if err == nil {
					switch resp := res.(type) {
					case *MsgVerifyInListResponse:
						t.Logf("ids in list: %v", resp)
					}
				}
				time.Sleep(1 * time.Second)
			}
			time.Sleep(3 * time.Second)

		})
	}
}
