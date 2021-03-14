package graph

type MsgWaitTag struct{}
type MsgWriteError struct{}
type MsgBalanceError struct {
	Value string
}
type MsgValidationTag struct {
	Value string
}
type MsgNewQr struct {
	Value string
}
type MsgRef struct {
	Version string
	Ruta    string
}
type MsgCount struct {
	Value int
}
