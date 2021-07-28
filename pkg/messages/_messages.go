package messages

type MsgPayment struct {
	UID  uint64
	Type uint
	Map  map[string]interface{}
}

type MsgWriteRequest struct {
	UID     uint64
	Updates map[string]interface{}
}
