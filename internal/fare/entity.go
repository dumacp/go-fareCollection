package fare

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	PLAIN = "P"
	INTEG = "I"
)

type Contidition struct {
	Days        []int  `json:"days"`
	InitTime    int    `json:"initTime"`
	EndTime     int    `json:"endTime"`
	TypeVehicle string `json:"typeVehicle"`
}

type Fare struct {
	ID           int                         `json:"id"`
	FarePolicyID string                      `json:"farePolicyId"`
	ProfileID    int                         `json:"profileID"`
	RouteID      int                         `json:"routeId"`
	ModeID       int                         `json:"modeId"`
	ItineraryID  int                         `json:"itineraryId"`
	ValidFrom    int64                       `json:"validFrom"`
	ValidTo      int64                       `json:"validTo"`
	TimeSpan     int64                       `json:"timeSpan"`
	Type         string                      `json:"type"`
	Fare         int                         `json:"fare"`
	Conditions   *Contidition                `json:"conditions"`
	Children     map[string]map[string][]int `json:"children"`
}

type FareNode struct {
	*Fare
	Kids   map[string]map[string][]*FareNode
	Parent *FareNode
}

// func (f *FareNode) AddChildren(childs ...*FareNode) {
// 	f.Kids = append(f.Kids, childs...)
// }

func (f *FareNode) FindChild(query *QueryFare) *FareNode {

	keysFrom := make([]string, 0)
	for k, _ := range f.Kids {
		keysFrom = append(keysFrom, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keysFrom)))

	for _, k := range keysFrom {
		from := f.Kids[k]
		keysTo := make([]string, 0)
		for k, _ := range from {
			keysTo = append(keysFrom, k)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(keysTo)))
		for _, k := range keysTo {
			to := from[k]
			for _, nextFare := range to {
				if query.VerifyFare(nextFare) {
					return nextFare
				}
			}
		}
	}
	return nil
}

type FareMap map[string][]*FareNode

func CreateTree(m map[int]*FareNode) FareMap {
	fm := make(map[string][]*FareNode)
	for _, node := range m {
		idString := &strings.Builder{}
		if node.ProfileID > 0 {
			idString.WriteString(fmt.Sprintf("%d", node.ProfileID))
		}
		if node.ModeID > 0 {
			idString.WriteString(fmt.Sprintf("-%d", node.ModeID))
		}
		if node.RouteID > 0 {
			idString.WriteString(fmt.Sprintf("-%d", node.RouteID))
		}
		if node.ItineraryID > 0 {
			idString.WriteString(fmt.Sprintf("%d", node.ItineraryID))
		}
		if _, ok := fm[idString.String()]; !ok {
			fm[idString.String()] = make([]*FareNode, 0)
		}
		fm[idString.String()] = append(fm[idString.String()], node)

		if len(node.Children) > 0 {
			if node.Kids == nil {
				node.Kids = make(map[string]map[string][]*FareNode)
			}
		}
		for kFrom, vFrom := range node.Children {
			if _, ok := node.Kids[kFrom]; !ok {
				node.Kids[kFrom] = make(map[string][]*FareNode)
			}
			for kTo, vTo := range vFrom {
				if _, ok := node.Kids[kFrom][kTo]; !ok {
					node.Kids[kFrom][kTo] = make([]*FareNode, 0)
				}
				for _, id := range vTo {
					if v, ok := m[id]; ok {
						node.Kids[kFrom][kTo] = append(node.Kids[kFrom][kTo], v)
					}
				}
			}
		}
	}
	return fm
}

func (fm FareMap) FindFare(query *QueryFare) *FareNode {

	for _, k := range query.KeyIndexes() {
		if fs, ok := fm[k]; ok {
			for _, f := range fs {
				if query.VerifyFare(f) {
					return f
				}
			}
		}
	}
	return nil
}

type QueryFare struct {
	ProfileID       int
	FromModeID      int
	FromRouteID     int
	FromItineraryID int
	ModeID          int
	RouteID         int
	ItineraryID     int
	Time            time.Time
}

func (query *QueryFare) VerifyFare(f *FareNode) bool {
	if f.ValidFrom > 0 && f.ValidFrom > query.Time.Unix() {
		return false
	}
	if f.ValidTo > 0 && f.ValidTo < query.Time.Unix() {
		return false
	}
	if query.ProfileID != 0 && query.ProfileID != f.ProfileID {
		return false
	}
	if query.ModeID != 0 && f.ModeID != 0 && query.ModeID != f.ModeID {
		return false
	}
	if query.RouteID != 0 && f.RouteID != 0 && query.RouteID != f.RouteID {
		return false
	}
	if query.ItineraryID != 0 && f.ItineraryID != 0 && query.ItineraryID != f.ItineraryID {
		return false
	}

	if f.Conditions != nil {
		validDay := false
		for _, day := range f.Conditions.Days {
			if day == int(query.Time.Weekday()) {
				validDay = true
				break
			}
		}
		if !validDay {
			return false
		}
		minutes := 60*query.Time.Hour() + query.Time.Minute()
		if minutes > f.Conditions.EndTime {
			return false
		}
		if minutes < f.Conditions.InitTime {
			return false
		}
	}

	return true
}

func (query *QueryFare) KeyIndexes() []string {
	idxes := make([]string, 0)
	idString := &strings.Builder{}
	if query.ProfileID > 0 {
		idString.WriteString(fmt.Sprintf("%d", query.ProfileID))
		idxes = append(idxes, idString.String())
	}
	if query.ModeID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.ModeID))
		idxes = append(idxes, idString.String())
	}
	if query.RouteID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.RouteID))
		idxes = append(idxes, idString.String())
	}
	if query.ItineraryID > 0 {
		idString.WriteString(fmt.Sprintf("%d", query.ItineraryID))
		idxes = append(idxes, idString.String())
	}
	reverse := make([]string, len(idxes))
	for i := range idxes {
		reverse = append(reverse, idxes[len(idxes)-1-i])
	}
	return reverse
}

// {
// 	"id": 10, // nestedId => paymentMediumCode
// 	"farePolicyId": "uuid",

// 	"profileId": 120, // profile.paymentMediumCode
// 	"modeId": 46, // mode.paymentMediumCode
// 	"routeId": 56, // route.paymentMediumCode
// 	"itineraryId": 66, // itinerary.paymentMediumCode
// 	"validFrom": 101010, // timestamp
// 	"validTo": 101010, // timestamp, can be null
// 	"timeSpan": 60000, // integration timespan

// 	"type": "P", // P=PLAIN  I=INTEG
// 	"fare": 2200, // amount to debit from card,
// 	"conditions": { // can be null
// 	  "days": [
// 		0, //Sunday
// 		1, // monday
// 		3, // Wensday
// 		5, // Friday
// 		6 // Saturday
// 	  ],
// 	  "initTime": 0,// minutes from 00:00
// 	  "endTime": 1080, // minutes from 00:00
// 	  "typeVehicle": "BUS"
// 	},
// 	"children": { //mode/route/itinerary with paymentMediumCode
// 	  "<mode>-<route>-<itinerary>": [
// 		12,
// 		34,
// 		13
// 	  ],
// 	  "<mode2>-<route2>": [
// 		55
// 	  ],
// 	  "<mode4>": [
// 		60
// 	  ]
// 	}
//   }
