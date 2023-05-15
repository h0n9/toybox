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
	setSubscribeCh    setSubscribeCh
	deleteSubscribeCh deleteSubscribeCh

	msgSubscribeCh SubscribeCh
	subscription   *pubsub.Subscription
	subscribers    map[string]SubscribeCh
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

		setSubscribeCh:    make(setSubscribeCh),
		deleteSubscribeCh: make(deleteSubscribeCh),

		msgSubscribeCh: make(SubscribeCh),
		subscription:   subscription,
		subscribers:    make(map[string]SubscribeCh),
	}
	box.wg.Add(1)
	go func() {
		defer box.wg.Done()
		var (
			msgSubscribe []byte

			setSubscribe    setSubscribe
			deleteSubscribe deleteSubscribe
		)
		for {
			select {
			case msgSubscribe = <-box.msgSubscribeCh:
				for _, subscribeCh := range box.subscribers {
					subscribeCh <- msgSubscribe
				}
			case setSubscribe = <-box.setSubscribeCh:
				_, exist := box.subscribers[setSubscribe.subscriberID]
				if exist {
					setSubscribe.errCh <- fmt.Errorf("%s is already subscribing", setSubscribe.subscriberID)
					continue
				}
				box.subscribers[setSubscribe.subscriberID] = setSubscribe.subscribeCh
				setSubscribe.errCh <- nil
			case deleteSubscribe = <-box.deleteSubscribeCh:
				_, exist := box.subscribers[deleteSubscribe.subscriberID]
				if !exist {
					deleteSubscribe.errCh <- fmt.Errorf("%s is not subscribing", <-deleteSubscribe.errCh)
					continue
				}
				delete(box.subscribers, deleteSubscribe.subscriberID)
				deleteSubscribe.errCh <- nil
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
			box.msgSubscribeCh <- msg.GetData()
		}
	}()
	return &box, nil
}

func (box *Box) Publish(data []byte) error {
	return box.topic.Publish(box.ctx, data)
}

func (box *Box) Subscribe(subscriberID string) (SubscribeCh, error) {
	var (
		subscribeCh = make(SubscribeCh)
		errCh       = make(chan error)
	)
	defer close(errCh)

	box.setSubscribeCh <- setSubscribe{
		subscriberID: subscriberID,
		subscribeCh:  subscribeCh,

		errCh: errCh,
	}
	err := <-errCh
	if err != nil {
		close(subscribeCh)
		return nil, err
	}

	return subscribeCh, nil
}

func (box *Box) StopSubscription(subscriberID string) error {
	var (
		errCh = make(chan error)
	)
	defer close(errCh)

	box.deleteSubscribeCh <- deleteSubscribe{
		subscriberID: subscriberID,

		errCh: errCh,
	}
	err := <-errCh
	if err != nil {
		return err
	}

	return nil
}
