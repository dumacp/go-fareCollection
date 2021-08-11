package usostransporte

type UsoTransporte struct {
	ID                    string      `json:"id"`
	PaymentMediumTypeCode int         `json:"paymentMediumCode"`
	PaymentMediumId       int         `json:"paymentMediumId"`
	MediumID              uint64      `json:"mediumId"`
	FareCode              int         `json:"fareCode"`
	RawDataPrev           interface{} `json:"rawDataPrev"`
	RawDataAfter          interface{} `json:"rawDataAfter"`
	Error                 *Error      `json:"error"`
	Coord                 string      `json:"coord"`
	// CountTrySend          int     `json:"trysend,omitempty"`
}

// type RawData interface{}

type Error struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
	Desc string `json:"Desc"`
	Addr string `json:"Addr"`
}
