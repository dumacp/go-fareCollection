package lists

type MsgTick struct{}

type MsgGetListsInDB struct{}

type MsgGetLists struct{}

type MsgGetListById struct {
	ID string
}
type MsgSetList struct {
	Data []byte
}

type MsgWatchList struct {
	ID string
}

type MsgVerifyInList struct {
	ListID string
	ID     []int64
}
type MsgVerifyInListResponse struct {
	ListID string
	ID     []int64
}
