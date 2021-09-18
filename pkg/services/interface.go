package services

import "github.com/dumacp/go-fareCollection/messages"

type Service interface {
	//Start
	Start()
	Stop()
	Restart()
	Status() *messages.StatusResponse

	// Info(ctx actor.Context, pid *actor.PID) (*messages.IgnitionStateResponse, error)
	// EventsSubscription(ctx actor.Context, pid *actor.PID) (*messages.IgnitionEventsSubscriptionAck, error)
}
