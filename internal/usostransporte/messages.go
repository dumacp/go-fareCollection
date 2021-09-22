package usostransporte

import "github.com/dumacp/go-fareCollection/internal/lock"

type MsgTick struct {
}

type MsgUso struct {
	Data *UsoTransporte
}
type MsgLock struct {
	Data *lock.Lock
}
type MsgSubscribe struct {
}
type MsgPublish struct{}
type MsgGetInDB struct{}
type MsgErrorDB struct {
	Error string
}
type MsgVerifyDB struct {
}
type MsgOkDB struct {
}
type MsgRequestStatus struct {
}
type MsgStatus struct {
	State bool
}
