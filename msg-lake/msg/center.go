package msg

import pubsub "github.com/libp2p/go-libp2p-pubsub"

type Center struct {
	ps    *pubsub.PubSub
	boxes map[string]*Box
}

func NewCenter(ps *pubsub.PubSub) *Center {
	return &Center{
		ps:    ps,
		boxes: make(map[string]*Box),
	}
}

func (center *Center) Join(topicID, subscriberID string) (SubscribeCh, error) {
	return nil, nil
}

func (center *Center) Leave(topicID, subscriberID string) error {
	return nil
}
