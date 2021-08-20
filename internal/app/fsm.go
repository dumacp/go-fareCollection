package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/dumacp/go-fareCollection/internal/buzzer"
	"github.com/dumacp/go-fareCollection/internal/graph"
	"github.com/dumacp/go-fareCollection/internal/picto"
	"github.com/dumacp/go-logs/pkg/logs"
	"github.com/looplab/fsm"
)

const (
	sStop           = "sStop"
	sStart          = "sStart"
	sDetectTag      = "sDetectTag"
	sValidationCard = "sValidationCard"
	sValidationQR   = "sValidationQR"
	sError          = "sError"
)

const (
	eStarted       = "eStarted"
	eCardDetected  = "eTagDetected"
	eCardValidated = "eCardValidated"
	eQRValidated   = "eQRValidated"
	eWait          = "eWait"
	eError         = "eError"
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

func (a *Actor) newFSM(callbacks fsm.Callbacks) {

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
			logs.LogBuild.Printf("FSM APP, state src: %s, state dst: %s", e.Src, e.Dst)
		},
		beforeEvent(eCardValidated): func(e *fsm.Event) {
			a.lastTime = time.Now()
			value := ""
			if e.Args != nil && len(e.Args) > 0 {
				switch v := e.Args[0].(type) {
				case int:
					value = FormatSaldo(v)
				case string:
					value = v
				}
				// if v, ok := e.Args[0].(int); ok {
				// 	// value = fmt.Sprintf("$%.02f", float64(v))
				// 	value = FormatSaldo(v)
				// }
			}
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerGood{})
			a.ctx.Send(a.pidPicto, &picto.MsgPictoOK{})
			a.ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: value})
			a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.inputs})

		},
		beforeEvent(eQRValidated): func(e *fsm.Event) {
			a.lastTime = time.Now()
			a.inputs++
			value := ""
			if e.Args != nil && len(e.Args) > 0 {
				if v, ok := e.Args[0].(string); ok {
					value = v
				}
			}
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerGood{})
			a.ctx.Send(a.pidPicto, &picto.MsgPictoOK{})
			a.ctx.Send(a.pidGraph, &graph.MsgValidationQR{Value: value})
			a.ctx.Send(a.pidGraph, &graph.MsgCount{Value: a.inputs})

		},
		enterState(sDetectTag): func(e *fsm.Event) {
			a.ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
			a.ctx.Send(a.pidPicto, &picto.MsgPictoOFF{})
		},
		beforeEvent(eError): func(e *fsm.Event) {
			a.lastTime = time.Now()
			if e.Args != nil && len(e.Args) > 0 {
				switch v := e.Args[0].(type) {
				case error:
					var err *ErrorShowInScreen
					if errors.As(v, &err) {
						a.ctx.Send(a.pidGraph, &graph.MsgError{Value: err.Value})
					}
				}
			}
			a.ctx.Send(a.pidPicto, &picto.MsgPictoNotOK{})
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})
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
		sStop,
		fsm.Events{
			{Name: eStarted, Src: []string{sStop}, Dst: sStart},
			{Name: eCardValidated, Src: []string{sDetectTag, sError}, Dst: sValidationCard},
			{Name: eQRValidated, Src: []string{sDetectTag, sError}, Dst: sValidationQR},
			{Name: eError, Src: []string{sDetectTag}, Dst: sError},
			{Name: eWait, Src: []string{
				sStart,
				sValidationQR,
				sValidationCard,
				sError,
			}, Dst: sDetectTag},
		},
		callbacksfsm,
	)
	a.fmachine = rfsm
}

func (a *Actor) RunFSM() {

	f := a.fmachine
	lastState := ""

	for {
		if err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					logs.LogError.Println("Recovered in \"startfsm() app\", ", r)
					switch x := r.(type) {
					case string:
						err = errors.New(x)
					case error:
						err = x
					default:
						err = errors.New("unknown panic")
					}
					time.Sleep(3 * time.Second)
				}
			}()

			//TODO: change!!!
			if f.Current() != sStop {
				f.SetState(sStart)
			}

			for {

				if lastState != f.Current() {
					lastState = f.Current()
					logs.LogBuild.Printf("current state app: %s", f.Current())
				}

				switch f.Current() {
				case sStart:
					a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerGood{})
					a.ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
					f.Event(eWait)
				case sDetectTag:
				case sValidationCard:
					if time.Now().Add(-time.Duration(a.timeout) * time.Millisecond).After(a.lastTime) {
						a.fmachine.Event(eWait)
						break
					}
				case sValidationQR:
					if time.Now().Add(-time.Duration(a.timeout) * time.Millisecond).After(a.lastTime) {
						a.fmachine.Event(eWait)
						break
					}
				case sError:
					if time.Now().Add(-10 * time.Second).After(a.lastTime) {
						a.fmachine.Event(eWait)
						break
					}
				}
				time.Sleep(300 * time.Millisecond)

			}
		}(); err != nil {
			logs.LogError.Println(err)
		}
	}
}
