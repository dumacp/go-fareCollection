package fare

import (
	"fmt"
	"strings"
	"time"
)

type QueryFare struct {
	LastFare        []*FareNode
	LastTimePlain   time.Time
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
	// log.Println("0 verify")
	if f.ValidFrom > 0 && f.ValidFrom > query.Time.Unix() {
		return false
	}
	// log.Println("1 verify")
	if f.ValidTo > 0 && f.ValidTo < query.Time.Unix() {
		return false
	}
	// log.Println("2 verify")
	if query.ProfileID != 0 && query.ProfileID != f.ProfileID {
		return false
	}
	// log.Println("3 verify")
	if query.ModeID != 0 && f.ModeID != 0 && query.ModeID != f.ModeID {
		return false
	}
	// log.Println("4 verify")
	if query.RouteID != 0 && f.RouteID != 0 && query.RouteID != f.RouteID {
		return false
	}
	// log.Println("5 verify")
	if query.ItineraryID != 0 && f.ItineraryID != 0 && query.ItineraryID != f.ItineraryID {
		return false
	}
	// log.Println("6 verify")

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
	// log.Printf("END verify: %+v", f)

	return true
}

func (query *QueryFare) KeyIndexes() []string {
	idxes := make([]string, 0)
	idString := &strings.Builder{}
	if query.ModeID > 0 {
		idString.WriteString(fmt.Sprintf("%d", query.ModeID))
		idxes = append(idxes, idString.String())
	}
	if query.RouteID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.RouteID))
		idxes = append(idxes, idString.String())
	}
	if query.ItineraryID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.ItineraryID))
		idxes = append(idxes, idString.String())
	}
	reverse := make([]string, len(idxes))
	for i := range idxes {
		reverse = append(reverse, idxes[len(idxes)-1-i])
	}
	return reverse
}

func (query *QueryFare) KeyIndexesFrom() []string {
	idxes := make([]string, 0)
	idString := &strings.Builder{}
	if query.FromModeID > 0 {
		idString.WriteString(fmt.Sprintf("%d", query.FromModeID))
		idxes = append(idxes, idString.String())
	}
	if query.FromRouteID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.FromRouteID))
		idxes = append(idxes, idString.String())
	}
	if query.FromItineraryID > 0 {
		idString.WriteString(fmt.Sprintf("-%d", query.FromItineraryID))
		idxes = append(idxes, idString.String())
	}
	reverse := make([]string, 0)
	for i := range idxes {
		reverse = append(reverse, idxes[len(idxes)-1-i])
	}
	return reverse
}

func (query *QueryFare) KeyIndexesWithProfile() []string {
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
		idString.WriteString(fmt.Sprintf("-%d", query.ItineraryID))
		idxes = append(idxes, idString.String())
	}
	reverse := make([]string, len(idxes))
	for i := range idxes {
		reverse = append(reverse, idxes[len(idxes)-1-i])
	}
	return reverse
}
