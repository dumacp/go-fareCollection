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

type Parameters struct {
	ID               string   `json:"id"`
	Timestamp        int64    `json:"timestamp"`
	PaymentMode      uint     `json:"paymentMode"`
	PaymentRoute     uint     `json:"paymentRoute"`
	PaymentItinerary uint     `json:"paymentItinerary"`
	RestrictiveList  []string `json:"restrictiveList"`
	Timeout          int      `json:"timeout"`
}

func (p *Parameters) FromPlatform(params *PlatformParameters) *Parameters {
	p.ID = params.ID
	p.PaymentMode = params.Mode()
	p.RestrictiveList = params.RestrictiveList()
	p.Timeout = params.Timeout()
	p.Timestamp = params.Timestamp
	return p
}

func (p *PlatformParameters) Mode() uint {
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

	return uint(res)
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
