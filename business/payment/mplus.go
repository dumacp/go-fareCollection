package payment

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/dumacp/go-fareCollection/crosscutting/logs"
)

// type Cardv1 interface {
// 	Name() string
// 	DocID() string
// 	CardID() int
// 	Perfil() int
// 	Version() int
// 	PMR() int
// 	AC() int
// 	Bloqueo() bool
// 	FechaBloque() time.Time
// 	Saldo() int
// 	saldoBackup() int
// 	Seq() int
// 	FechaRecarga() time.Time
// 	SeqRecarga() int
// 	ValorRecarga() int
// 	TipoRecarga()
// 	FechaValidez() time.Time
// }

const (
	cost = 2400
)

func ValidationTag(tag map[string]interface{}, ruta int, devID int) (map[string]interface{}, error) {

	saldo1, ok := tag["saldo"]
	if !ok {
		return nil, errors.New("saldo field is empty")
	}
	saldo2, ok := tag["saldoBackup"]
	if !ok {
		return nil, errors.New("saldoBk field is empty")
	}

	saldo := saldo1.(int32)
	if saldo > saldo2.(int32) {
		saldo = saldo2.(int32)
	}

	if saldo < cost {
		return nil, fmt.Errorf("error: %w", &ErrorBalanceValue{Balance: float64(saldo)})
	}
	newTag := make(map[string]interface{})
	if saldo1.(int32) > saldo2.(int32) {
		newTag["saldo"] = saldo1.(int32) - (saldo1.(int32) - saldo2.(int32)) - cost
		tag["newSaldo"] = saldo1.(int32) - cost
	} else {
		newTag["saldoBackup"] = saldo2.(int32) - (saldo2.(int32) - saldo1.(int32)) - cost
		tag["newSaldo"] = saldo2.(int32) - cost
	}

	v := tag["seq"].(int32) + 1
	newTag["seq"] = v

	hist, err := History(tag)
	if err != nil {
		return nil, err
	}
	for _, v := range hist {
		logs.LogBuild.Printf("map hist: %+v", v)
	}
	if len(hist) < 1 {
		return nil, errors.New("hist error")
	}

	hist[0].Date = time.Now()
	hist[0].DevID = devID
	hist[0].Perfil = 0
	hist[0].Route = ruta
	hist[0].Seq = 0
	hist[0].Valor = cost
	hist[0].WalletType = 1

	for k, v := range hist[0].ToMapping() {
		newTag[k] = v
	}

	return newTag, nil
}

type Hist struct {
	Index      int
	Date       time.Time
	Valor      int
	DevID      int
	Route      int
	Perfil     int
	Seq        int
	WalletType int
}

func (h *Hist) ToMapping() map[string]interface{} {

	prefix := fmt.Sprintf("hist%d_", h.Index)

	mapp := make(map[string]interface{})
	mapp[fmt.Sprintf("%s%s", prefix, "time")] = uint32(h.Date.Unix())
	mapp[fmt.Sprintf("%s%s", prefix, "valor")] = int32(h.Valor)
	mapp[fmt.Sprintf("%s%s", prefix, "iddev")] = uint32(h.DevID)
	mapp[fmt.Sprintf("%s%s", prefix, "ruta")] = uint32(h.Route)
	mapp[fmt.Sprintf("%s%s", prefix, "perfil")] = uint16(h.Perfil)
	mapp[fmt.Sprintf("%s%s", prefix, "seqi")] = uint16(h.Seq)
	mapp[fmt.Sprintf("%s%s", prefix, "wallett")] = uint16(h.WalletType)

	return mapp
}

func History(tag map[string]interface{}) ([]*Hist, error) {

	re, err := regexp.Compile(`hist([0-9])_(.+)`)
	if err != nil {
		return nil, err
	}
	hists := make(map[string]*Hist)
	for k, v := range tag {
		res := re.FindStringSubmatch(k)
		if len(res) > 2 && len(res[2]) > 0 {
			// logs.LogBuild.Printf("regexp hist: %v", res)
			ind := res[1]
			key := res[2]
			if _, ok := hists[ind]; !ok {
				indx, _ := strconv.Atoi(ind)
				hists[ind] = &Hist{Index: indx}
			}
			switch key {
			case "time":
				hists[ind].Date = time.Unix(int64(v.(uint32)), int64(0))
			case "valor":
				hists[ind].Valor = int(v.(int32))
			case "iddev":
				hists[ind].DevID = int(v.(uint32))
			case "ruta":
				hists[ind].Route = int(v.(uint32))
			case "seqi":
				hists[ind].Seq = int(v.(uint16))
			case "wallett":
				hists[ind].WalletType = int(v.(uint16))
			}
		}
	}

	result := make([]*Hist, 0)
	for _, v := range hists {
		result = append(result, v)
	}

	sort.SliceStable(result, func(i, j int) bool { return result[i].Date.Before(result[j].Date) })
	return result, nil
}
