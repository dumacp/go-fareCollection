package app

import (
	"fmt"
	"log"

	"github.com/looplab/fsm"
)

const (
	sStart     = "sStart"
	sDetectTag = "sDetectTag"
	sReadCard  = "sReadTag"
	sReadQR    = "sReadQR"
	sWriteCard = "sWriteCard"
	// sErrorTag       = "sErrorTag"
	sValidationCard = "sValidationCard"
	sValidationQR   = "sValidationQR"
	sWaitScreen     = "sWaitScreen"
)

const (
	eStarted       = "eStarted"
	eCardDetected  = "eTagDetected"
	eQRDetected    = "eQRDetected"
	eCardRead      = "eCardRead"
	eQRRead        = "eQRRead"
	eCardValidated = "eCardValidated"
	eQRValidated   = "eQRValidated"
	// eErrorQR        = "eErrorQR"
	// eErrorCardRead  = "eErrorCardRead"
	// eErrorCardWrite = "eErrorCardWrite"
	eScreenWait = "eScreenWait"
	eStep       = "eStep"
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

func newFSM(callbacks *fsm.Callbacks) *fsm.FSM {

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

	for k, v := range *callbacks {
		callbacksfsm[k] = v
	}

	rfsm := fsm.NewFSM(
		sStart,
		fsm.Events{
			{Name: eStarted, Src: []string{sStart}, Dst: sDetectTag},
			{Name: eCardDetected, Src: []string{sDetectTag}, Dst: sReadCard},
			{Name: eQRDetected, Src: []string{sDetectTag}, Dst: sReadQR},
			{Name: eCardRead, Src: []string{sReadCard}, Dst: sValidationCard},
			{Name: eCardValidated, Src: []string{sValidationCard}, Dst: sWriteCard},
			{Name: eScreenWait, Src: []string{
				sWriteCard,
				sValidationQR,
				sValidationCard,
			}, Dst: sWaitScreen},
		},
		callbacksfsm,
	)
	return rfsm
}

func RunFSM(f *fsm.FSM) {

	stepPending := false
	// dataTag := make(map[string]interface{})

	for {

		switch f.Current() {
		case sStart:
		case sDetectTag:

		case sReadCard:
			if stepPending {
				stepPending = false
				f.Event(eStep)
			}
			f.Event(eCardDetected)
		case sValidationCard:

		}

	}

}
