package fare

import (
	"reflect"
	"testing"
	"time"

	"github.com/dumacp/go-fareCollection/internal/itinerary"
)

func TestFareNode_FindFare(t *testing.T) {

	itsMap := make(itinerary.ItineraryMap)
	itsMap[110] = &itinerary.Itinerary{
		PaymentMediumCode:      110,
		RoutePaymentMediumCode: 11,
		ModePaymentMediumCode:  1,
	}
	itsMap[210] = &itinerary.Itinerary{
		PaymentMediumCode:      210,
		RoutePaymentMediumCode: 21,
		ModePaymentMediumCode:  2,
	}
	itsMap[111] = &itinerary.Itinerary{
		PaymentMediumCode:      111,
		RoutePaymentMediumCode: 11,
		ModePaymentMediumCode:  1,
	}

	fareBus := &FareNode{
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
	}

	fareMetro := &FareNode{
		ID:           2200,
		FarePolicyID: "2200",
		ProfileID:    1,
		ValidFrom:    10000000,
		ValidTo:      0,
		TimeSpan:     60 * 60,
		Type:         PLAIN,
		ModeID:       2,
		Fare:         3000,
		Children:     nil,
	}

	fareBus_BusRuta11 := &FareNode{
		ID:           1011,
		FarePolicyID: "1011",
		ProfileID:    1,
		ValidFrom:    10000000,
		ValidTo:      0,
		TimeSpan:     60 * 60,
		Type:         INTEG,
		ModeID:       1,
		RouteID:      11,
		Fare:         1200,
		Children: map[string]map[string][]int{
			"1-11": {
				"2": []int{2011},
			},
		},
	}

	fareBus_Metro := &FareNode{
		ID:           2000,
		FarePolicyID: "2000",
		ProfileID:    1,
		ValidFrom:    10000000,
		ValidTo:      0,
		TimeSpan:     30 * 60,
		Type:         INTEG,
		ModeID:       2,
		RouteID:      21,
		Fare:         400,
		Children:     nil,
	}

	fareBus_MetroLineA := &FareNode{
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
	}

	fareBusRuta11_Metro := &FareNode{
		ID:           2011,
		FarePolicyID: "2011",
		ProfileID:    1,
		ValidFrom:    10000000,
		ValidTo:      0,
		TimeSpan:     30 * 60,
		Type:         INTEG,
		ModeID:       2,
		RouteID:      21,
		Fare:         100,
		Children:     nil,
	}

	fareBusIti111_Metro := &FareNode{
		ID:           2111,
		FarePolicyID: "2111",
		ProfileID:    1,
		ValidFrom:    10000000,
		ValidTo:      0,
		TimeSpan:     60 * 60,
		Type:         INTEG,
		ModeID:       2,
		Fare:         100,
		Children:     nil,
	}

	Fares := map[int]*FareNode{
		fareBus.ID:             fareBus,
		fareMetro.ID:           fareMetro,
		fareBus_BusRuta11.ID:   fareBus_BusRuta11,
		fareBus_Metro.ID:       fareBus_Metro,
		fareBus_MetroLineA.ID:  fareBus_MetroLineA,
		fareBusRuta11_Metro.ID: fareBusRuta11_Metro,
		fareBusIti111_Metro.ID: fareBusIti111_Metro,
	}

	fMap := CreateTree(Fares)

	type args struct {
		query *QueryFare
	}

	tests := []struct {
		name string
		args args
		want *FareNode
	}{
		{
			name: "test1",
			args: args{
				query: &QueryFare{
					ProfileID:       1,
					RouteID:         11,
					ItineraryID:     110,
					ModeID:          1,
					LastFare:        nil,
					LastTimePlain:   time.Time{},
					FromModeID:      0,
					FromRouteID:     0,
					FromItineraryID: 0,
					Time:            time.Now(),
				},
			},
			want: fareBus,
		},
		{
			name: "test2",
			args: args{
				query: &QueryFare{
					ProfileID:       1,
					ModeID:          1,
					RouteID:         11,
					FromRouteID:     11,
					FromItineraryID: 110,
					FromModeID:      1,
					LastFare:        []*FareNode{fareBus},
					LastTimePlain:   time.Now().Add(-60 * time.Second),
					Time:            time.Now(),
				},
			},
			want: fareBus_BusRuta11,
		},
		{
			name: "test3",
			args: args{
				query: &QueryFare{
					ProfileID:       1,
					ModeID:          2,
					RouteID:         21,
					FromRouteID:     11,
					FromItineraryID: 110,
					FromModeID:      1,
					LastFare:        []*FareNode{fareBus, fareBus_BusRuta11},
					LastTimePlain:   time.Now().Add(-60 * time.Minute),
					Time:            time.Now(),
				},
			},
			want: fareBusRuta11_Metro,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fMap.FindFare(tt.args.query); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FareNode.FindFare() = %+v, want %v", got, tt.want)
			}
		})
	}
}
