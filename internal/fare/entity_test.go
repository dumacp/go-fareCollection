package fare

import (
	"reflect"
	"testing"
)

func TestFareNode_FindChild(t *testing.T) {

	fareBus := &FareNode{
		Fare: &Fare{
			ID:           1001,
			FarePolicyID: "1001",
			ProfileID:    1,
			ValidFrom:    10000000,
			ValidTo:      0,
			TimeSpan:     60 * 60,
			Type:         PLAIN,
			ModeID:       1,
			Fare:         2000,
			Children: map[string]map[string][]int{
				"1": {
					"1-11": []int{1011, 1012},
					"2-21": []int{2001, 2000},
					"2":    []int{2000}},
				"1-11": {
					"2": []int{2011},
				},
				"1-11-111": {
					"2": []int{2111},
				},
			},
		},
	}

	fareBus_Metro := &FareNode{
		Fare: &Fare{
			ID:           2000,
			FarePolicyID: "2000",
			ProfileID:    1,
			ValidFrom:    10000000,
			ValidTo:      0,
			TimeSpan:     60 * 60,
			Type:         INTEG,
			ModeID:       2,
			RouteID:      21,
			Fare:         400,
			Children:     nil,
		},
	}

	fareBus_MetroLineA := &FareNode{
		Fare: &Fare{
			ID:           2001,
			FarePolicyID: "2001",
			ProfileID:    1,
			ValidFrom:    10000000,
			ValidTo:      0,
			TimeSpan:     60 * 60,
			Type:         INTEG,
			ModeID:       2,
			RouteID:      21,
			Fare:         300,
			Children:     nil,
		},
	}

	fareBusRuta11_Metro := &FareNode{
		Fare: &Fare{
			ID:           2001,
			FarePolicyID: "2001",
			ProfileID:    1,
			ValidFrom:    10000000,
			ValidTo:      0,
			TimeSpan:     60 * 60,
			Type:         INTEG,
			ModeID:       2,
			RouteID:      21,
			Fare:         300,
			Children:     nil,
		},
	}

	fareBusIti111_Metro := &FareNode{
		Fare: &Fare{
			ID:           2111,
			FarePolicyID: "2111",
			ProfileID:    1,
			ValidFrom:    10000000,
			ValidTo:      0,
			TimeSpan:     60 * 60,
			Type:         INTEG,
			ModeID:       1,
			RouteID:      11,
			Fare:         100,
			Children:     nil,
		},
	}

	Fares := map[int]*FareNode{
		fareBus.ID:             fareBus,
		fareBus_Metro.ID:       fareBus_Metro,
		fareBus_MetroLineA.ID:  fareBus_MetroLineA,
		fareBusRuta11_Metro.ID: fareBusRuta11_Metro,
		fareBusIti111_Metro.ID: fareBusIti111_Metro,
	}

	fMap := CreateTree(Fares)

	type fields struct {
		FareID      int
		ItineraryID int
		ProfileID   int
	}
	type args struct {
		query *QueryFare
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FareNode
	}{
		{
			name: "test1",
			fields: fields{
				FareID:      0,
				ItineraryID: 0,
				ProfileID:   1},
			args: args{
				query: &QueryFare{
					ProfileID:       1,
					FromItineraryID: 0,
					ModeID:          1,
				},
			},
		},
		{
			name: "test2",
			fields: fields{
				FareID:      1001,
				ItineraryID: 0,
				ProfileID:   1},
			args: args{
				query: &QueryFare{
					ProfileID:       1,
					FromItineraryID: 0,
					ModeID:          1,
					RouteID:         11,
				},
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q1 := &QueryFare{
				ProfileID: tt.fields.ProfileID,
				ModeID:    1,
			}
			f := fMap.FindFare(q1)
			if got := f.FindChild(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FareNode.FindChild() = %v, want %v", got, tt.want)
			}
		})
	}
}
