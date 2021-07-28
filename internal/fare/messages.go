package fare

type MsgGetFarePolicies struct{}
type MsgGetFare struct {
	LastFarePolicies map[int64]int // key: timestamp (seconds), value: FarePolicyID
	ProfileID        int
	ItineraryID      int
	RouteID          int
	ModeID           int
	FromItineraryID  int
}
type MsgResponseFare struct {
	Fare         int
	FarePolicyID int
}
type MsgError struct {
	Err string
}
