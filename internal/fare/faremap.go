package fare

import (
	"fmt"
	"log"
	"strings"
)

type FareMap map[string][]*FareNode

func CreateTree(m map[int]*FareNode) FareMap {
	fm := make(map[string][]*FareNode)
	for _, node := range m {
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
		if node.Type != PLAIN {
			continue
		}
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
			idString.WriteString(fmt.Sprintf("-%d", node.ItineraryID))
		}
		if _, ok := fm[idString.String()]; !ok {
			fm[idString.String()] = make([]*FareNode, 0)
		}
		fm[idString.String()] = append(fm[idString.String()], node)
	}
	return fm
}

func (fm FareMap) FindPlainFare(query *QueryFare) *FareNode {
	for _, k := range query.KeyIndexesWithProfile() {
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

func (fm FareMap) FindFare(query *QueryFare) *FareNode {

	if len(query.LastFare) <= 0 {
		return fm.FindPlainFare(query)
	}
	if query.LastFare[0].Type != PLAIN {
		return fm.FindPlainFare(query)
	}
	if query.FromModeID <= 0 {
		return fm.FindPlainFare(query)
	}
	if query.FromRouteID <= 0 && query.FromItineraryID > 0 {
		return fm.FindPlainFare(query)
	}

	// if query.LastFare[0].TimeSpan > 0 &&
	// 	((time.Duration(query.LastFare[0].TimeSpan) * time.Second) < query.Time.Sub(query.LastTimePlain)) {
	// 	return fm.FindPlainFare(query)
	// }

	lastFare := query.LastFare[len(query.LastFare)-1]
	if fare := lastFare.FindChild(query); fare != nil {
		return fare
	}

	log.Println("without CHILD")

	return fm.FindPlainFare(query)
}
