package app

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func FormatSaldo(data int) string {
	// res := make([]byte, 0)

	// if data > 1000_000 {
	// 	v := data / 1000_1000
	// 	res = append(res, []byte(fmt.Sprintf("%d'", v))...)
	// }
	// if data%1000_000/1000 > 999 {
	// 	v := data / 1000
	// 	res = append(res, []byte(fmt.Sprintf("%d'", v))...)
	// }
	p := message.NewPrinter(language.LatinAmericanSpanish)
	return p.Sprintf("%d", data)

	// return fmt.Sprintf("%s", res)
}
