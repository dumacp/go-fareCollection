package app

type MsgTagDetected struct {
	UID uint64
}
type MsgTagRead struct {
	Data map[string]interface{}
}
type MsgTagWriteError struct {
	Block int
}
type MsgQRRead struct {
	Data map[string]interface{}
}
