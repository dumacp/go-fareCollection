package app

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"math/rand"
	"time"
)

const (
	keyQr   = "5mYqX4wY4YfgtGSt"
	timeout = 30
)

var ErrorCipher = errors.New("Error in CIPHER")

type QrCode struct {
	Route         string `json:"r"`
	Device        string `json:"d"`
	Pin           string `json:"p"`
	TransactionID int32  `json:"t"`
}

func (a *Actor) TickQR(ch <-chan int) {

	//TODO: add chquit
	tick1 := time.NewTimer(5 * time.Second)
	defer tick1.Stop()

	for {
		select {
		case <-ch:
			if a.ctx != nil {
				a.ctx.Send(a.ctx.Self(), &MsgNewRand{Value: int(NewCode())})
			}
			if !tick1.Stop() {
				select {
				case <-tick1.C:
				default:
				}
			}
			tick1.Reset(timeout * time.Second)
		case <-tick1.C:
			if a.ctx != nil {
				a.ctx.Send(a.ctx.Self(), &MsgNewRand{Value: int(NewCode())})
			}
			if !tick1.Stop() {
				select {
				case <-tick1.C:
				default:
				}
			}
			tick1.Reset(timeout * time.Second)
		}
	}
}

func NewCode() int32 {

	rand.Seed(time.Now().UnixNano())
	v1 := 12000 + rand.Int31n(10000)
	rand.Seed(time.Now().UnixNano())
	v2 := rand.Int31n(10000)

	return v1 + v2
}

func DecodeQR(data []byte) ([]byte, error) {
	key := []byte(keyQr)
	iv := make([]byte, 16)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	dcrypt := cipher.NewCBCDecrypter(block, iv)
	for len(data)%dcrypt.BlockSize() != 0 {
		return nil, ErrorCipher
	}
	dts := make([]byte, len(data))
	dcrypt.CryptBlocks(dts, data)

	return dts, nil
}
