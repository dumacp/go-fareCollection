package parameters

import (
	"strconv"
	"strings"
)

type PlatformParameters struct {
	ID                  string            `json:"id"`
	ApplicationIds      []string          `json:"applicationIds"`
	GroupID             string            `json:"groupId"`
	HasErrors           bool              `json:"hasErrors"`
	NewParameterMapHash string            `json:"newParameterMapHash"`
	ParameterMapId      string            `json:"parameterMapId"`
	Props               map[string]string `json:"props"`
	Timestamp           int64             `json:"timestamp"`
	VehicleId           string            `json:"vehicleId"`
}

type ConfigParameters struct {
	Expiration       int64 `json:"exp"`
	VerificationCode int64 `json:"c"`
	PaymentItinerary int   `json:"id"`
}

type AppParameters struct {
	Seq     uint `json:"seq"`
	Inputs  int  `json:"inputs"`
	Outputs int  `json:"outputs"`
}

type Parameters struct {
	ID               string   `json:"id"`
	Timestamp        int64    `json:"timestamp"`
	PaymentMode      int      `json:"paymentMode"`
	PaymentRoute     int      `json:"paymentRoute"`
	PaymentItinerary int      `json:"paymentItinerary"`
	RestrictiveList  []string `json:"restrictiveList"`
	Timeout          int      `json:"timeout"`
	Inputs           int      `json:"inputs"`
	Outputs          int      `json:"outputs"`
	Seq              uint     `json:"seq"`
	DevSerial        int      `json:"devSerial"`
	UrlQR            string   `json:"urlQR"`
	KeyQr            int      `json:"keyQR"`
}

func (p *Parameters) FromPlatform(params *PlatformParameters) *Parameters {
	p.ID = params.ID
	p.PaymentMode = params.Mode()
	p.RestrictiveList = params.RestrictiveList()
	p.Timeout = params.Timeout()
	p.DevSerial = params.Serial()
	p.Timestamp = params.Timestamp
	p.UrlQR = params.URLQR()
	p.KeyQr = params.KeyQR()
	return p
}

func (p *Parameters) FromConfig(params *ConfigParameters) *Parameters {
	// if params.Expiration > time.Now().UnixNano()/1000_000 {
	// 	return p
	// }
	p.PaymentItinerary = params.PaymentItinerary
	return p
}

func (p *Parameters) FromApp(params *AppParameters) *Parameters {
	// if params.Expiration > time.Now().UnixNano()/1000_000 {
	// 	return p
	// }
	p.Inputs = params.Inputs
	p.Outputs = params.Outputs
	p.Seq = params.Seq
	return p
}

func (p *PlatformParameters) Mode() int {
	if p.Props == nil {
		return 0
	}
	modess, ok := p.Props["MODE"]
	if !ok {
		return 0
	}

	res, err := strconv.Atoi(modess)
	if err != nil {
		return 0
	}

	return int(res)
}

func (p *PlatformParameters) RestrictiveList() []string {
	if p.Props == nil {
		return nil
	}
	listss, ok := p.Props["PML_RESTRICTIVE"]
	if !ok {
		return nil
	}

	result := make([]string, 0)
	ss := strings.Split(listss, ",")
	for _, s := range ss {
		result = append(result, strings.TrimSpace(s))
	}
	return result
}

func (p *PlatformParameters) Timeout() int {
	if p.Props == nil {
		return 0
	}
	timeouts, ok := p.Props["TIMEOUT"]
	if !ok {
		return 0
	}

	res, err := strconv.Atoi(timeouts)
	if err != nil {
		return 0
	}

	return res
}

func (p *PlatformParameters) URLQR() string {
	if p.Props == nil {
		return ""
	}
	url, ok := p.Props["QR_URL"]
	if !ok {
		return ""
	}

	return url
}

func (p *PlatformParameters) KeyQR() int {
	if p.Props == nil {
		return 0
	}
	key, ok := p.Props["QR_SLOT_KEY"]
	if !ok {
		return 0
	}

	res, err := strconv.Atoi(key)
	if err != nil {
		return 0
	}

	return int(res)
}

func (p *PlatformParameters) Serial() int {
	if p.Props == nil {
		return 0
	}
	serial, ok := p.Props["DEV_SERIAL"]
	if !ok {
		return 0
	}

	res, err := strconv.Atoi(serial)
	if err != nil {
		return 0
	}

	return res
}

// curl --location --request GET 'https://fleet.nebulae.com.co/api/external-system-gateway/rest/dev-summary/MNP137' \
// --header 'Authorization: Basic bGVvbmFyZG8uZ3V0aWVycmV6OnVuby4yLnRyZXM='

// RESPUESTA:
// {
//     "_id": "MNP137",
//     "applicationIds": [],
//     "groupId": "ba30430a-05f8-43ab-8478-dc1d4af44029",
//     "hasErrors": false,
//     "newParameterMapHash": "b59efa218b49837f1be3ce244c32c55e",
//     "parameterMapId": "fb601af8-a7d1-4aca-8f44-5a3a873f76de",
//     "props": {},
//     "timestamp": 1626367969584,
//     "vehicleId": "4c27d9d7-0b78-4ea5-95d4-78765976e338",
//     "id": "MNP137"
// }
