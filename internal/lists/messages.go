package lists

type MsgTick struct{}

type MsgGetListsInDB struct{}

type MsgGetLists struct{}

type MsgGetListById struct {
	ID   string
	Code string
}
type MsgWatchList struct {
	ID string
}
type WatchList struct {
	ID                string
	PaymentMediumType string
	Version           float32
}

type MsgVerifyInList struct {
	ListID string
	ID     []int64
}
type MsgVerifyInListResponse struct {
	ListID string
	ID     []int64
}
