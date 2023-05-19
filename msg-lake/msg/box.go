package msg

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type Box struct {
	ctx context.Context
	wg  sync.WaitGroup

	topicID string
	topic   *pubsub.Topic

	// chans for operations
	setSubscriberCh    setSubscriberCh
	deleteSubscriberCh deleteSubscriberCh

	subscriberCh SubscriberCh
	subscription *pubsub.Subscription
	subscribers  map[string]SubscriberCh
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

		setSubscriberCh:    make(setSubscriberCh),
		deleteSubscriberCh: make(deleteSubscriberCh),

		subscriberCh: make(SubscriberCh),
		subscription: subscription,
		subscribers:  make(map[string]SubscriberCh),
	}
	box.wg.Add(1)
	go func() {
		defer box.wg.Done()
		var (
			msg []byte

			setSubscriber    setSubscriber
			deleteSubscriber deleteSubscriber
		)
		for {
			select {
			case msg = <-box.subscriberCh:
				for _, subscriberCh := range box.subscribers {
					subscriberCh <- msg
				}
			case setSubscriber = <-box.setSubscriberCh:
				_, exist := box.subscribers[setSubscriber.subscriberID]
				if exist {
					setSubscriber.errCh <- fmt.Errorf("%s is already subscribing", setSubscriber.subscriberID)
					continue
				}
				box.subscribers[setSubscriber.subscriberID] = setSubscriber.subscriberCh
				setSubscriber.errCh <- nil
			case deleteSubscriber = <-box.deleteSubscriberCh:
				_, exist := box.subscribers[deleteSubscriber.subscriberID]
				if !exist {
					deleteSubscriber.errCh <- fmt.Errorf("%s is not subscribing", <-deleteSubscriber.errCh)
					continue
				}
				delete(box.subscribers, deleteSubscriber.subscriberID)
				deleteSubscriber.errCh <- nil
			}
		}
	}()

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
			box.subscriberCh <- msg.GetData()
		}
	}()
	return &box, nil
}

func (box *Box) Publish(data []byte) error {
	return box.topic.Publish(box.ctx, data)
}

func (box *Box) Subscribe(subscriberID string) (SubscriberCh, error) {
	var (
		subscriberCh = make(SubscriberCh)
		errCh        = make(chan error)
	)
	defer close(errCh)

	box.setSubscriberCh <- setSubscriber{
		subscriberID: subscriberID,
		subscriberCh: subscriberCh,

		errCh: errCh,
	}
	err := <-errCh
	if err != nil {
		close(subscriberCh)
		return nil, err
	}

	return subscriberCh, nil
}

func (box *Box) StopSubscription(subscriberID string) error {
	var (
		errCh = make(chan error)
	)
	defer close(errCh)

	box.deleteSubscriberCh <- deleteSubscriber{
		subscriberID: subscriberID,

		errCh: errCh,
	}
	err := <-errCh
	if err != nil {
		return err
	}

	return nil
}
