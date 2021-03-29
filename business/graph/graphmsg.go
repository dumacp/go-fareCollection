package graph

import (
	"encoding/json"
)

//MessageGraph type graph messages to send to JS screen
type MessageGraph struct {
	Typef string      `json:"type"`
	Value interface{} `json:"value"`
}

//Screen value type for type screen
type Screen struct {
	ID  int      `json:"id"`
	Msg []string `json:"msg"`
}

//Countinputs value type for type inputs
type Countinputs struct {
	Count int `json:"count"`
}

//ReferenceApp value type for type ref
type ReferenceApp struct {
	Appversion string `json:"appversion"`
	Refproduct string `json:"refproduct"`
	// Count      int    `json:"count"`
}

type Qrvalue struct {
	URL string `json:"url"`
}

const (
	Tittleqr          = "qr"
	Tittlecountinputs = "countinputs"
	Tittleref         = "ref"
	Tittlescreen      = "screen"
)

//Funcgraphmsg closure function to build graphs messages
func Funcgraphmsg(types string) func(interface{}) ([]byte, error) {

	typess := types

	return func(sc interface{}) ([]byte, error) {
		msg := new(MessageGraph)

		msg.Typef = typess

		msg.Value = sc

		sendMsg, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		return sendMsg, nil
	}
}

var funScreen = Funcgraphmsg(Tittlescreen)
var funCounts = Funcgraphmsg(Tittlecountinputs)
var funQR = Funcgraphmsg(Tittleqr)
var funRef = Funcgraphmsg(Tittleref)
