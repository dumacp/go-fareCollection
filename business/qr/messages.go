package qr

type MsgNewCodeQR struct {
	Value []byte
}
type MsgResponseCodeQR struct {
	Value int
}
