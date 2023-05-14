package msg

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type SubscribeCh chan []byte

type Box struct {
	ctx context.Context
	wg  sync.WaitGroup

	topicID string
	topic   *pubsub.Topic

	subscription *pubsub.Subscription
	subscribers  map[string]SubscribeCh
}

func NewBox(ctx context.Context, topicID string, topic *pubsub.Topic) (*Box, error) {
	subscription, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}
	box := Box{
		ctx: ctx,
		wg:  sync.WaitGroup{},

		topicID: topicID,
		topic:   topic,

		subscription: subscription,
		subscribers:  make(map[string]SubscribeCh),
	}
	box.wg.Add(1)
	go func() {
		defer box.wg.Done()
		defer subscription.Cancel()
		for {
			msg, err := subscription.Next(ctx)
			if err != nil {
				fmt.Println(err)
				return
			}
			data := msg.GetData()
			for _, subscriber := range box.subscribers {
				subscriber <- data
			}
		}
	}()
	return &box, nil
}

func (box *Box) Publish(data []byte) error {
	return box.topic.Publish(box.ctx, data)
}

func (box *Box) Subscribe(subscriberID string) (SubscribeCh, error) {
	subscribeCh, exist := box.subscribers[subscriberID]
	if exist {
		return subscribeCh, nil
	}
	subscribeCh = make(SubscribeCh)
	box.subscribers[subscriberID] = subscribeCh
	return subscribeCh, nil
}
