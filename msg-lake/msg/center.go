package msg

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type Center struct {
	ctx context.Context

	ps    *pubsub.PubSub
	boxes map[string]*Box
}

func NewCenter(ctx context.Context, ps *pubsub.PubSub) *Center {
	return &Center{
		ctx: ctx,

		ps:    ps,
		boxes: make(map[string]*Box),
	}
}

func (center *Center) Join(topicID, subscriberID string) (SubscribeCh, error) {
	box, exist := center.boxes[topicID]
	if !exist {
		topic, err := center.ps.Join(topicID)
		if err != nil {
			return nil, err
		}
		newBox, err := NewBox(center.ctx, topicID, topic)
		if err != nil {
			return nil, err
		}
		box = newBox
		center.boxes[topicID] = box
	}
	return box.Subscribe(subscriberID)
}

func (center *Center) Leave(topicID, subscriberID string) error {
	box, exist := center.boxes[topicID]
	if !exist {
		return fmt.Errorf("topic '%s' doesn't exist", topicID)
	}
	return box.StopSubscription(subscriberID)
}
