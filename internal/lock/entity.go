package lock

type Reason string

func (r Reason) String() string {
	return string(r)
}

const (
	ON_RESTRICTIVE_LIST Reason = "ON_RESTRICTIVE_LIST"
)

type Lock struct {
	ID                       string      `json:"transactionId"`
	DeviceID                 string      `json:"deviceId"`
	PaymentMediumTypeCode    string      `json:"paymentMediumTypeCode"`
	PaymentMediumId          string      `json:"paymentMediumId"`
	SamUuid                  string      `json:"samUuid"`
	MediumID                 string      `json:"mediumId"`
	RawDataPrev              interface{} `json:"rawDataPrev"`
	RawDataAfter             interface{} `json:"rawDataAfter"`
	Error                    *Error      `json:"error"`
	Coord                    string      `json:"coord"`
	TransactionTime          int64       `json:"timestamp"`
	Reason                   string      `json:"reason"`
	PaymentMediumListVersion float32     `json:"paymentMediumListVersion"`
	PaymentMediumListId      string      `json:"paymentMediumListId"`
}

type Error struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
	Desc string `json:"Desc"`
}
