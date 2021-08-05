package gps

import "time"

//MsgGPS GPRMC Data
type MsgGPS struct {
	Data []byte
}

type MsgGpsRaw struct {
	Data []byte
	Time time.Time
}

type MsgGetGps struct{}

type MsgSubscribe struct{}
