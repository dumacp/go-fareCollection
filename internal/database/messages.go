package database

type MsgOpenDB struct{}
type MsgOpenedDB struct{}
type MsgCloseDB struct{}
type MsgUpdateData struct {
	ID         string
	Data       []byte
	Database   string
	Collection string
}
type MsgDeleteData struct {
	ID         string
	Database   string
	Collection string
}
type MsgInsertData struct {
	ID         string
	Data       []byte
	Database   string
	Collection string
}
type MsgGetData struct {
	Database   string
	Collection string
	ID         string
}
type MsgQueryData struct {
	Database   string
	Collection string
	PrefixID   string
	Reverse    bool
}
type MsgQueryNext struct{}
type MsgQueryResponse struct {
	ID         string
	Database   string
	Collection string
	Data       []byte
}
type MsgAckPersistData struct {
	ID string
}
type MsgNoAckPersistData struct {
	Error string
}

type MsgAckGetData struct {
	Data []byte
}
type MsgNoAckGetData struct {
	Error string
}
