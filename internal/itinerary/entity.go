package itinerary

type Mode struct {
	ID                string `json:"id"`
	Active            bool   `json:"active"`
	OrganizationId    string `json:"organizationId"`
	PaymentMediumCode int    `json:"paymentMediumCode"`
	Name              string `json:"name"`
}

// {
// 	"active": true,
// 	"name": "Bus",
// 	"organizationId": "604812bb-ed5a-4a46-b076-104dbf0dc982",
// 	"id": "5a9d62dc-6439-40c5-bd57-9c6c69815289",
// 	"paymentMediumCode": 1
// }

type Route struct {
	ID                    string `json:"id"`
	Active                bool   `json:"active"`
	OrganizationId        string `json:"organizationId"`
	Name                  string `json:"name"`
	Code                  int    `json:"code"`
	AuthorityCode         string `json:"authorityCode"`
	DivipolCode           string `json:"divipolCode"`
	ModeID                string `json:"modeId"`
	PaymentMediumCode     int    `json:"paymentMediumCode"`
	ModePaymentMediumCode int    `json:"modePaymentMediumCode"`
}

// {
// 	"_id": "1c5dc64a-73f4-49ca-b971-208e69b66917",
// 	"active": true,
// 	"organizationId": "604812bb-ed5a-4a46-b076-104dbf0dc982",
// 	"name": "CENTRO - LA MOTA - CENTRO",
// 	"code": "5546",
// 	"authorityCode": "6646",
// 	"divipolCode": "",
// 	"modeId": "5a9d62dc-6439-40c5-bd57-9c6c69815289",
// 	"paymentMediumCode": 91,
// 	"id": "1c5dc64a-73f4-49ca-b971-208e69b66917",
// 	"modePaymentMediumCode": 1
// }

type Itinerary struct {
	ID                     string `json:"id"`
	Active                 bool   `json:"active"`
	Direction              string `json:"direction"`
	Name                   string `json:"name"`
	RouteID                string `json:"routeId"`
	PaymentMediumCode      int    `json:"paymentMediumCode"`
	ModePaymentMediumCode  int    `json:"modePaymentMediumCode"`
	RoutePaymentMediumCode int    `json:"routePaymentMediumCode"`
}

type ItineraryMap map[int]*Itinerary

// {
// 	"_id": "1719d8a9-eda4-447c-8d9e-4ca2d77de437",
// 	"active": true,
// 	"direction": "DEPART",
// 	"name": "CIRCULAR 300",
// 	"routeId": "80f85843-4419-40e4-ac13-1af3ed356343",
// 	"paymentMediumCode": 274,
// 	"id": "1719d8a9-eda4-447c-8d9e-4ca2d77de437",
// 	"modePaymentMediumCode": 1
// 	"routePaymentMediumCode": 91
// }
