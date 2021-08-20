package mplus

import (
	"testing"
)

func Test_mplus_AddBalance(t *testing.T) {

	data := map[string]interface{}{
		"AC": 0, "B": 0, "CT$": 287, "FV": 0, "HISTR_CK_1": 0, "HISTR_CK_2": 0, "HISTR_CT_1": 0,
		"HISTR_CT_2": 0, "HISTR_FT_1": 0, "HISTR_FT_2": 0, "HISTR_IDV_1": 0, "HISTR_IDV_2": 0,
		"HISTR_TT_1": 0, "HISTR_TT_2": 0, "HISTR_VT_1": 0, "HISTR_VT_2": 0, "HISTU_CK_1": 0,
		"HISTU_CK_2": 0, "HISTU_CK_3": 0, "HISTU_CK_4": 0, "HISTU_CK_5": 0, "HISTU_CK_6": 0,
		"HISTU_FID_1": 60, "HISTU_FID_2": 60, "HISTU_FID_3": 60, "HISTU_FID_4": 60, "HISTU_FID_5": 60,
		"HISTU_FID_6": 60, "HISTU_FT_1": 1629476142, "HISTU_FT_2": 1629476128,
		"HISTU_FT_3": 1629476131, "HISTU_FT_4": 1629476135, "HISTU_FT_5": 1629476138,
		"HISTU_FT_6": 1629476125, "HISTU_IDV_1": 1, "HISTU_IDV_2": 1, "HISTU_IDV_3": 1,
		"HISTU_IDV_4": 1, "HISTU_IDV_5": 1, "HISTU_IDV_6": 1, "HISTU_ITI_1": 294, "HISTU_ITI_2": 294,
		"HISTU_ITI_3": 294, "HISTU_ITI_4": 294, "HISTU_ITI_5": 294, "HISTU_ITI_6": 294,
		"NT": 423155169, "P": 1, "PMR": 0, "ST$": 1492000, "STB$": 1490000, "VL": 1,
	}

	type args struct {
		value int
	}
	tests := []struct {
		name      string
		args      args
		wantValue int
	}{
		// TODO: Add test cases.
		{
			name: "tes1",
			args: args{
				value: 3000,
			},
			wantValue: 1493000,
		},
		{
			name: "tes2",
			args: args{
				value: -3000,
			},
			wantValue: 1490000,
		},
	}
	p := ParseToPayment(12345, "TYPE1", data)
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if err := p.AddBalance(tt.args.value); err != nil {
				t.Errorf("mplus.AddBalance() error = %v", err)
			}
			t.Logf("updates: %+v", p.Updates())
			if p.Balance() != tt.wantValue {
				t.Errorf("mplus.AddBalance() error: balance = %v, want = %v", p.Balance(), tt.wantValue)
			}
		})
	}
}
