package fare

import (
	"log"
	"sort"
	"strings"
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

type FareNode struct {
	ID           int                               `json:"id"`
	FarePolicyID string                            `json:"farePolicyId"`
	ProfileID    int                               `json:"profileID"`
	RouteID      int                               `json:"routeId"`
	ModeID       int                               `json:"modeId"`
	ItineraryID  int                               `json:"itineraryId"`
	ValidFrom    int64                             `json:"validFrom"`
	ValidTo      int64                             `json:"validTo"`
	TimeSpan     int                               `json:"timeSpan"`
	Type         string                            `json:"type"`
	Fare         int                               `json:"fare"`
	Conditions   *Contidition                      `json:"conditions"`
	Children     map[string]map[string][]int       `json:"children"`
	Kids         map[string]map[string][]*FareNode `json:"-"`
}

func (f *FareNode) FindChild(query *QueryFare) *FareNode {

	keysFrom := make(prefix, 0)
	for k, _ := range f.Kids {
		// log.Printf("keys: %v", k)
		keysFrom = append(keysFrom, k)
	}
	sort.Sort(sort.Reverse(keysFrom))
	log.Printf("%d, keys FROM: %v", f.ID, keysFrom)

	for _, indexFrom := range query.KeyIndexesFrom() {
		log.Printf("indexFROM: %s", indexFrom)
		for _, k := range keysFrom {
			if !strings.HasPrefix(indexFrom, k) {
				continue
			}
			log.Printf("keyFrom, queryFrom: %s, %s", k, indexFrom)
			from := f.Kids[k]
			keysTo := make(prefix, 0)
			for k := range from {
				keysTo = append(keysTo, k)
			}
			sort.Sort(sort.Reverse(keysTo))
			log.Printf("keys TO: %v", keysTo)
			for _, indexTo := range query.KeyIndexes() {
				for _, k := range keysTo {
					if !strings.HasPrefix(indexTo, k) {
						continue
					}
					log.Printf("keyTo, queryTo: %s, %s", k, indexTo)
					to := from[k]
					for _, nextFare := range to {
						log.Printf("verify Fare: %+v", nextFare)
						if query.VerifyFare(nextFare) {
							return nextFare
						}
					}
				}
			}
		}
	}
	return nil
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
