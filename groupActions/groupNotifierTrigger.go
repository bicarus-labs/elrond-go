package groupActions

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/groupActions/groupTypes"
)

// Subscriber gets notified upon event
type Subscriber interface {
	ReceiveNotification(triggerID string, header data.HeaderHandler, stage groupTypes.TriggerStage)
	IsInterfaceNil() bool
}

type groupNotifierTrigger struct {
	triggerID   string
	notifyOrder uint32
	subscriber  Subscriber
}

func newGroupNotifierTrigger(triggerID string, notifyOrder uint32, subscriber Subscriber) (*groupNotifierTrigger, error) {
	if len(triggerID) == 0 {
		return nil, errInvalidTriggerID
	}
	if check.IfNil(subscriber) {
		return nil, errNilSubscriber
	}

	return &groupNotifierTrigger{
		triggerID:   triggerID,
		notifyOrder: notifyOrder,
		subscriber:  subscriber,
	}, nil
}

// Prepare -
func (nt *groupNotifierTrigger) Prepare(hdr data.HeaderHandler, _ data.BodyHandler) {
	nt.subscriber.ReceiveNotification(nt.triggerID, hdr, groupTypes.Prepare)
}

// Action -
func (nt *groupNotifierTrigger) Action(hdr data.HeaderHandler) {
	nt.subscriber.ReceiveNotification(nt.triggerID, hdr, groupTypes.Action)
}

// NotifyOrder -
func (nt *groupNotifierTrigger) NotifyOrder() uint32 {
	return nt.notifyOrder
}

// IsInterfaceNil returns true if the receiver is nil and true otherwise
func (nt *groupNotifierTrigger) IsInterfaceNil() bool {
	return nt == nil
}
