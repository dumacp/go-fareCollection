package token

import "errors"

func (t *token) ApplyFare(data interface{}) (interface{}, error) {

	switch value := data.(type) {
	case []int:
		for _, v := range value {
			if t.pin == v {
				return v, nil
			}
		}
		return 0, errors.New("pin is invalid")
	}
	return 0, errors.New("pin parse is invalid")
}

func (t *token) Balance() int {
	return int(t.id)
}

func (t *token) FareID() uint {
	return 0
}
