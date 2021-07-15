package database

import "time"

type MsgPersistData struct {
	ID        string
	Data      []byte
	Indexes   []string
	TimeStamp time.Time
	Database  string
}

type MsgAckPersistData struct {
	ID     string
	Succes bool
	Error  string
}
