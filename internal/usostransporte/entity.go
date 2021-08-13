package usostransporte

type UsoTransporte struct {
	ID                     string      `json:"transactionId"`
	DeviceID               string      `json:"deviceId"`
	PaymentMediumTypeCode  string      `json:"paymentMediumTypeCode"`
	TerminalTransactionSeq int         `json:"terminalTransactionSeq"`
	PaymentMediumId        string      `json:"paymentMediumId"`
	SamUuid                string      `json:"samUuid"`
	MediumID               string      `json:"mediumId"`
	FareCode               int         `json:"fareCode"`
	RawDataPrev            interface{} `json:"rawDataPrev"`
	RawDataAfter           interface{} `json:"rawDataAfter"`
	Error                  *Error      `json:"error"`
	Coord                  string      `json:"coord"`
	TransactionType        string      `json:"transactionType"`
	// CountTrySend          int     `json:"trysend,omitempty"`
}

// type RawData interface{}

type Error struct {
	Code int    `json:"Code"`
	Name string `json:"Name"`
	Desc string `json:"Desc"`
}
