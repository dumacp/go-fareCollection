package qr

import (
	"errors"
	"math/rand"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// const (
// 	keyQr = "5mYqX4wY4YfgtGSt"
// 	// timeout = 30
// )

var ErrorCipher = errors.New("Error in CIPHER")

type QrCode struct {
	Route         string `json:"r"`
	Device        string `json:"d"`
	Pin           string `json:"p"`
	TransactionID int32  `json:"t"`
}

func tickQR(ctx actor.Context, quit, ch <-chan int) {

	timeout := 30 * time.Second
	rootctx := ctx.ActorSystem().Root
	self := ctx.Self()
	t1 := time.NewTimer(5 * time.Second)
	defer t1.Stop()

	for {
		select {
		case <-ch:
			rootctx.Send(self, &MsgNewRand{Value: int(NewCode())})
			if !t1.Stop() {
				select {
				case <-t1.C:
				default:
				}
			}
			t1.Reset(timeout)
		case <-t1.C:
			rootctx.Send(self, &MsgNewRand{Value: int(NewCode())})
			if !t1.Stop() {
				select {
				case <-t1.C:
				default:
				}
			}
			t1.Reset(timeout)
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

// func DecodeQR(data []byte) ([]byte, error) {
// 	key := []byte(keyQr)
// 	iv := make([]byte, 16)
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	dcrypt := cipher.NewCBCDecrypter(block, iv)
// 	for len(data)%dcrypt.BlockSize() != 0 {
// 		return nil, ErrorCipher
// 	}
// 	dts := make([]byte, len(data))
// 	dcrypt.CryptBlocks(dts, data)

// 	return dts, nil
// }
