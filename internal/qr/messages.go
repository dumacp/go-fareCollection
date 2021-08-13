package qr

type MsgNewCodeQR struct {
	Value []byte
}
type MsgResponseCodeQR struct {
	Value []byte
}

type MsgNewRand struct {
	Value int
}
