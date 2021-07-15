package app

import (
	"reflect"
	"testing"
)

func TestDecodeQR(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecodeQR(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeQR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeQR() = %v, want %v", got, tt.want)
			}
		})
	}
}
