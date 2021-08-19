package qr

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
	"github.com/tarm/serial"
)

const (
	portSerial = "/dev/ttyQR"
	portSpeed  = 9600
)

const (
	sStart = "sStart"
	sOpen  = "sOpen"
	sClose = "sClose"
	sRead  = "sRead"
)

const (
	eStarted = "eStarted"
	eOpened  = "eOpened"
	eClosed  = "eClosed"
	eRead    = "eRead"
	eError   = "eError"
)

// var keyQR = []byte{0x06, 0xB3, 0x0E, 0x65, 0x72, 0x3E, 0x3C, 0x96, 0x48, 0x8E, 0xD4, 0x05, 0xF1, 0x24, 0x2E, 0x88}

// func beforeEvent(event string) string {
// 	return fmt.Sprintf("before_%s", event)
// }
// func enterState(state string) string {
// 	return fmt.Sprintf("enter_%s", state)
// }
// func leaveState(state string) string {
// 	return fmt.Sprintf("leave_%s", state)
// }

func NewFSM(callbacks fsm.Callbacks) *fsm.FSM {

	callbacksfsm := fsm.Callbacks{
		"before_event": func(e *fsm.Event) {
			if e.Err != nil {
				// log.Println(e.Err)
				e.Cancel(e.Err)
			}
		},
		"leave_state": func(e *fsm.Event) {
			if e.Err != nil {
				// log.Println(e.Err)
				e.Cancel(e.Err)
			}
		},
		"enter_state": func(e *fsm.Event) {
			logs.LogBuild.Printf("FSM QR, state src: %s, state dst: %s", e.Src, e.Dst)
		},

		// "leave_closed": func(e *fsm.Event) {
		// },
		// "before_verify": func(e *fsm.Event) {
		// },
		// "enter_closed": func(e *fsm.Event) {
		// },
	}

	for k, v := range callbacks {
		callbacksfsm[k] = v
	}

	rfsm := fsm.NewFSM(
		sClose,
		fsm.Events{
			{Name: eOpened, Src: []string{sClose}, Dst: sOpen},
			{Name: eRead, Src: []string{sOpen, sRead}, Dst: sRead},
			{Name: eError, Src: []string{sOpen, sRead}, Dst: sClose},
		},
		callbacksfsm,
	)
	return rfsm
}

func (a *Actor) RunFSM(quit <-chan int) {

	if err := func() (errr error) {
		defer func() {
			if r := recover(); r != nil {
				logs.LogError.Println("Recovered in \"startfsm() app\", ", r)
				switch x := r.(type) {
				case string:
					errr = errors.New(x)
				case error:
					errr = x
				default:
					errr = errors.New("unknown panic")
				}
			}
		}()

		var port *serial.Port
		var reader *bufio.Reader
		lastState := ""
		lastRead := ""
		lastTime := time.Now().Add(300 * time.Second)

		for {
			select {
			case <-quit:
				return nil
			default:
			}
			if lastState != a.fmachine.Current() {
				lastState = a.fmachine.Current()
				logs.LogBuild.Printf("current state QR reader: %s", a.fmachine.Current())
			}

			switch a.fmachine.Current() {
			case sStart:
			case sOpen:
				config := &serial.Config{
					Name: portSerial,
					Baud: 9600,
					//ReadTimeout: time.Second * 3,
				}
				var err error
				succ := false
				for _, baud := range []int{9600, 19200, 9600} {
					config.Baud = baud
					port, err = serial.OpenPort(config)
					if err != nil {
						time.Sleep(2 * time.Second)
						continue
					}
					if _, err = port.Write([]byte("?")); err != nil {
						time.Sleep(2 * time.Second)
						continue
					}
					buff := make([]byte, 16)
					if _, err = port.Read(buff); err != nil {
						if errors.Is(err, io.EOF) && bytes.Contains(buff, []byte("!")) {
							succ = true
							break
						}
						break
					}
					if bytes.Contains(buff, []byte("!")) {
						succ = true
						break
					}
				}
				if !succ {
					return err
				}
				reader = bufio.NewReader(port)
				a.fmachine.Event(eRead)
			case sRead:

				if err := func() error {
					v, err := reader.ReadBytes(0x0D)
					if err != nil {
						if !errors.Is(err, io.EOF) {
							a.fmachine.Event(eError)
							return fmt.Errorf("QR read error: %s", err)
						}
						return nil
					}
					// code := base64.RawStdEncoding.EncodeToString(v)
					if lastRead == string(v) {
						if time.Now().Add(-5 * time.Second).Before(lastTime) {
							return nil
						}
					}
					lastTime = time.Now()
					lastRead = string(v)
					logs.LogBuild.Printf("QR read: %s", v)
					data, err := base64.StdEncoding.DecodeString(string(v))
					if err != nil {
						return fmt.Errorf("QR error: %w", err)
					}

					a.ctx.Request(a.ctx.Parent(), &MsgNewCodeQR{Value: data})
					return nil
				}(); err != nil {
					logs.LogError.Println(err)
				}
			case sClose:
			}
			time.Sleep(10 * time.Millisecond)

		}
	}(); err != nil {
		logs.LogError.Printf("QR error: %s", err)
		//TODO send MsgFatal
	}

}
