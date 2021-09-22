package mediums

type MsgTick struct {
}
type MsgGetPaymentMedium struct {
}
type MsgGetPaymentMediumByID struct {
	ID string
}
type MsgPaymentMedium struct {
	Data *PaymentMedium
}
type MsgSubscribePaymentMedium struct {
}
type MsgPublishPaymentMedium struct{}
type MsgGetInDB struct{}
type MsgRequestStatus struct {
}
type MsgStatus struct {
	State bool
}
