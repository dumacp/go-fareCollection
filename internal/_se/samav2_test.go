package se

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dumacp/smartcard/nxp/mifare/samav2"
	"github.com/dumacp/smartcard/pcsc"
)

func Test_enableKeys(t *testing.T) {

	ctx, err := pcsc.NewContext()
	if err != nil {
		t.Fatal(err)
	}
	rs, err := ctx.ListReaders()
	if err != nil {
		t.Fatal(err)
	}
	var r pcsc.Reader

	for _, v := range rs {
		if strings.Contains(v, "SAM") {
			r = pcsc.NewReader(ctx, v)
			break
		}
	}
	if r == nil {
		t.Fatalf("reader SAM not found")
	}

	card, err := r.ConnectSamCard()
	if err != nil {
		t.Fatal(err)
	}

	sam := samav2.SamAV2(card)

	keyMaster := make([]byte, 16)
	if _, err := sam.AuthHostAV2(keyMaster, 0, 0, 0); err != nil {
		t.Fatal(err)
	}
	t.Log("Auth!!!")

	type args struct {
		se samav2.SamAv2
	}
	tests := []struct {
		name    string
		args    args
		want    map[int]int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				se: sam,
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := enableKeys(tt.args.se)
			if (err != nil) != tt.wantErr {
				t.Errorf("enableKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("enableKeys() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
