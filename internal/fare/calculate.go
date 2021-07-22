package fare

import (
	"sort"
)

//P=PLAIN  I=INTEG lastFarePolicies is map with key = timestamp and value = fareId
func FindFareWithMap(lastFarePolicies map[int]int,
	query *QueryFare,
	farePolicies map[int]*FareNode, fareMap FareMap) *FareNode {

	keysTimestamp := make([]int, 0)
	// valuesId := make([]int, 0)
	for _, v := range lastFarePolicies {
		keysTimestamp = append(keysTimestamp, v)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(keysTimestamp)))

	lastFare := farePolicies[lastFarePolicies[keysTimestamp[0]]]
	for _, timestamp := range keysTimestamp {
		fare := farePolicies[lastFarePolicies[timestamp]]
		if fare.Type == PLAIN {
			//timespan in time
			if int64(fare.TimeSpan) < query.Time.Unix()-int64(keysTimestamp[0]) {
				return lastFare.FindChild(query)
			}
		}

	}

	// for k, v := range lastFarePolicies {
	// 	farePolicies[k]
	// }
	return nil
}
