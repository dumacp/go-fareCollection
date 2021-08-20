package app

// type MsgTagDetected struct {
// 	UID uint64
// }
// type MsgTagRead struct {
// 	UID  uint64
// 	Data map[string]interface{}
// }
// type MsgTagWriteError struct {
// 	Block int
// }
// type MsgQRRead struct {
// 	Data map[string]interface{}
// }

type MsgTick struct{}
type MsgWriteErrorVerify struct{}
type MsgWriteAppParamas struct{}
