package fare

type MsgGetFarePolicies struct{}
type MsgGetFare struct {
	LastFarePolicies map[int]int64 // key: FarePolicyId, value: timestamp (seconds)
}
