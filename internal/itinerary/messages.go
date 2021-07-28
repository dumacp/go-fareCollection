package itinerary

type MsgGetMap struct{}
type MsgMap struct {
	Data ItineraryMap
}
type MsgTick struct{}
type MsgGetModes struct{}
type MsgGetRoutes struct{}
type MsgGetItinerary struct{}
