package se

import (
	"fmt"
	"time"

	"github.com/dumacp/go-hsm/pkg/logs"
	"github.com/looplab/fsm"
)

const (
	sOpen      = "sOpen"
	sClose     = "sClose"
	sRestart   = "sRestart"
	sWaitEvent = "sWait"
)

const (
	eOpenCmd = "eOpenCmd"
	eOpened  = "eOpened"
	eClosed  = "eClosed"
	eError   = "eError"
)

// func beforeEvent(event string) string {
// 	return fmt.Sprintf("before_%s", event)
// }

func enterState(state string) string {
	return fmt.Sprintf("enter_%s", state)
}

// func leaveState(state string) string {
// 	return fmt.Sprintf("leave_%s", state)
// }

func (a *samActor) initFSM() {

	calls := fsm.Callbacks{
		"enter_state": func(e *fsm.Event) {
			logs.LogBuild.Printf("FSM SAM state Src: %v, state Dst: %v", e.Src, e.Dst)
		},
		"leave_state": func(e *fsm.Event) {
			if e.Err != nil {
				e.Cancel(e.Err)
			}
		},
		"before_event": func(e *fsm.Event) {
			if e.Err != nil {
				e.Cancel(e.Err)
			}
		},
		enterState(sOpen): func(e *fsm.Event) {
			a.behavior.Become(a.WaitState)
			go a.fm.Event(eOpened)
		},
		enterState(sClose): func(e *fsm.Event) {
			a.behavior.Become(a.CloseState)
			// a.ctx.Request(a.db.PID(), &database.MsgCloseDB{})
		},
		enterState(sRestart): func(e *fsm.Event) {
			time.Sleep(10 * time.Second)
			a.behavior.Become(a.CloseState)
			a.ctx.Send(a.ctx.Self(), &MsgOpen{})
		},
	}

	f := fsm.NewFSM(
		sClose,
		fsm.Events{
			{Name: eOpenCmd, Src: []string{sClose, sRestart}, Dst: sOpen},
			{Name: eOpened, Src: []string{sOpen}, Dst: sWaitEvent},
			{Name: eClosed, Src: []string{sOpen, sWaitEvent}, Dst: sClose},
			// {Name: eError, Src: []string{sOpen, sWaitEvent}, Dst: sRestart},
		},
		calls,
	)
	a.fm = f
}
