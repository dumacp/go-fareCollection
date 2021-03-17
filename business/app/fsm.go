package app

import (
	"fmt"
	"log"
	"time"

	"github.com/dumacp/go-fareCollection/business/buzzer"
	"github.com/dumacp/go-fareCollection/business/graph"
	"github.com/looplab/fsm"
)

const (
	sStart     = "sStart"
	sDetectTag = "sDetectTag"
	sValidationCard = "sValidationCard"
	sValidationQR   = "sValidationQR"
	sError = "sError"
)

const (
	eStarted       = "eStarted"
	eCardDetected  = "eTagDetected"
	eCardValidated = "eCardValidated"
	eQRValidated   = "eQRValidated"
	eWait = "eWait"
	eError       = "eError"
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

func (a *Actor)newFSM(callbacks *fsm.Callbacks) {

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
		beforeEvent(eCardValidated): func(e *fsm.Event) {
			a.inputs++
			value := ""
			if e.Args != nil && len(e.Args) > 0 {
				if v, ok := e.Args[0].(float64)
				value = fmt.Sprintf("$%.02f", v)
			}
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerGood{})
			a.ctx.Send(a.pidGraph, &graph.MsgValidationTag{Value: value})
			
		}
		enterState(sDetectTag): func(e *fsm.Event) {	
			a.ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
		},		
		ebeforeEvent(eError): func(e *fsm.Event) {
			if e.Args != nil && len(e.Args) > 0 {
				if v, ok := e.Args[0].([]string) {
					a.ctx.Send(a.pidGraph, &graph.MsgError{Value: v})
				}
			}	
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerBad{})		
		}

		// "leave_closed": func(e *fsm.Event) {
		// },
		// "before_verify": func(e *fsm.Event) {
		// },
		// "enter_closed": func(e *fsm.Event) {
		// },
	}

	for k, v := range *callbacks {
		callbacksfsm[k] = v
	}

	rfsm := fsm.NewFSM(
		sStart,
		fsm.Events{
			{Name: eStarted, Src: []string{sStart}, Dst: sDetectTag},
			{Name: eCardValidated, Src: []string{sDetectTag}, Dst: sValidationCard},
			{Name: eQRValidated, Src: []string{sDetectTag}, Dst: sValidationQR},
			{Name: eWait, Src: []string{
				sValidationQR,
				sValidationCard,
				sError,
			}, Dst: sDetectTag},
		},
		callbacksfsm,
	)
	a.fmachine = rfsm
}

func (a *Actor)RunFSM() {

	stepPending := false
	// dataTag := make(map[string]interface{})

	for {

		switch f.Current() {
		case sStart:
			a.ctx.Send(a.pidBuzzer, &buzzer.MsgBuzzerGood{})
			a.ctx.Send(a.pidGraph, &graph.MsgWaitTag{})
		case sDetectTag:
		case sValidationCard:
			if time.Now().Add(5 * time.Second).After(a.lastTime) {
				a.fmachine.Event(eWait)
				break
			}
		case sValidationQR:
			if time.Now().Add(5 * time.Second).After(a.lastTime) {
				a.fmachine.Event(eWait)
				break
			}
		case sError:
			if time.Now().Add(5 * time.Second).After(a.lastTime) {
				a.fmachine.Event(eWait)
				break
			}
		}
		time.Sleep(300 * time.Millisecond)

	}

}
