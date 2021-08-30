package qr

type MsgNewCodeQR struct {
	Value []byte
}
type MsgResponseCodeQR struct {
	Value  []byte
	SamUid string
}

type MsgNewRand struct {
	Value int
}
