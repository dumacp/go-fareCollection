package usostransporte

type MsgTick struct {
}

type MsgUso struct {
	Data *UsoTransporte
}
type MsgSubscribe struct {
}
type MsgPublish struct{}
type MsgGetInDB struct{}
