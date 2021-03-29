package database

import "time"

type MsgPersistData struct {
	ID        int64
	Data      []byte
	TimeStamp time.Time
	Database  string
}
