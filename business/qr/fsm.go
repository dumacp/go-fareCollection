package qr

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/dumacp/go-fareCollection/crosscutting/logs"
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

func beforeEvent(event string) string {
	return fmt.Sprintf("before_%s", event)
}
func enterState(state string) string {
	return fmt.Sprintf("enter_%s", state)
}
func leaveState(state string) string {
	return fmt.Sprintf("leave_%s", state)
}

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
			log.Printf("FSM APP, state src: %s, state dst: %s", e.Src, e.Dst)
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

func RunFSM(ctx *actor.Context, chQuit chan int, f *fsm.FSM) {

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
					errr = errors.New("Unknown panic")
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
			case _, _ = <-chQuit:
				return nil
			default:
			}

			if lastState != f.Current() {
				lastState = f.Current()
				logs.LogBuild.Printf("current state QR reader: %s", f.Current())
			}

			switch f.Current() {
			case sStart:
			case sOpen:
				config := &serial.Config{
					Name: portSerial,
					Baud: 9600,
					//ReadTimeout: time.Second * 3,
				}
				var err error
				succ := false
				for range []int{1, 2, 3} {
					port, err = serial.OpenPort(config)
					if err != nil {
						time.Sleep(2 * time.Second)
						continue
					}
					succ = true
					break
				}
				if !succ {
					return err
				}
				reader = bufio.NewReader(port)
				f.Event(eRead)
			case sRead:

				v, err := reader.ReadBytes(0x0D)
				if err != nil {
					logs.LogError.Printf("QR read error: %s", err)
					if !errors.Is(err, io.EOF) {
						f.Event(eError)
						break
					}
					break
				}
				// code := base64.RawStdEncoding.EncodeToString(v)
				if lastRead == string(v) {
					if time.Now().Add(-5 * time.Second).Before(lastTime) {
						break
					}
				}
				lastTime = time.Now()
				lastRead = string(v)
				logs.LogBuild.Printf("QR read: %s", v)
				data, err := base64.StdEncoding.DecodeString(string(v))
				if err != nil {
					logs.LogWarn.Printf("QR error: %s", err)
					break
				}
				(*ctx).Send((*ctx).Self(), &MsgNewCodeQR{Value: data})
			case sClose:
			}
			time.Sleep(10 * time.Millisecond)

		}
	}(); err != nil {
		logs.LogError.Printf("QR error: %s", err)
		//TODO send MsgFatal
	}

}
