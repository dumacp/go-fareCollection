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

func (m *MsgGetFare) GetPriority() int8 {
	return 7
}

type MsgFare struct {
	Fare         int
	FarePolicyID int
	ItineraryID  int
	DeviceID     int
}
type MsgError struct {
	Err string
}

type MsgTick struct{}

type MsgGetFareInDB struct{}

type MsgRequestStatus struct {
}
type MsgStatus struct {
	State bool
}
