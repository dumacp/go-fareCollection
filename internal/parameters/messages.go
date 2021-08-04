package parameters

type MsgTick struct {
}

type MsgGetParameters struct {
}

type MsgParameters struct {
	Data *Parameters
}
type MsgSubscribe struct {
}
type MsgPublish struct{}
type MsgGetInDB struct{}
