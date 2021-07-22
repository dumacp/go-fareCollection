package itinerary

type Itinerary struct {
	ID      int
	RouteID int
	ModeID  int
}

type ItineraryMap map[int]*Itinerary
