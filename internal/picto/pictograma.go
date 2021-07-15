package picto

import "io/ioutil"

//Pictograma type to represent picto map
type Pictograma int

const (
	PictogreenON Pictograma = iota
	PictogreenOFF
	PictoredON
	PictoredOFF
)

var pictoON = []byte("1")
var pictoOFF = []byte("0")

const (
	pictogreen = "/sys/class/leds/picto-go-gren/brightness"
	pictored   = "/sys/class/leds/picto-stop-red/brightness"
)

//PictoInit init picto dev files
func PictoInit() error {

	pictogreenbad := "/sys/class/leds/picto-stop-gren/brightness"
	pictoredbad := "/sys/class/leds/picto-go-red/brightness"

	if err := ioutil.WriteFile(pictogreenbad, pictoOFF, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(pictogreen, pictoOFF, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(pictoredbad, pictoOFF, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(pictored, pictoOFF, 0644); err != nil {
		return err
	}
	return nil
}

//PictoFunc func to send pictogram
func PictoFunc(picto Pictograma) error {

	switch picto {
	case PictogreenOFF:
		if err := ioutil.WriteFile(pictogreen, pictoOFF, 0644); err != nil {
			return err
		}
	case PictogreenON:
		if err := ioutil.WriteFile(pictogreen, pictoON, 0644); err != nil {
			return err
		}
	case PictoredOFF:
		if err := ioutil.WriteFile(pictored, pictoOFF, 0644); err != nil {
			return err
		}
	case PictoredON:
		if err := ioutil.WriteFile(pictored, pictoON, 0644); err != nil {
			return err
		}
	}
	return nil
}
