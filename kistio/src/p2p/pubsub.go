package p2p

import (
	"fmt"

	libp2pPubsub "github.com/libp2p/go-libp2p-pubsub"
)

func (n *Node) JoinTopic(name string) (*libp2pPubsub.Topic, error) {
	err := n.checkTopic(name)
	if err != nil {
		return nil, err
	}
	topic, err := n.pubsub.Join(name)
	if err != nil {
		return nil, err
	}
	n.topics[name] = topic
	return topic, err
}

func (n *Node) checkTopic(name string) error {
	_, exist := n.topics[name]
	if exist {
		return fmt.Errorf("found '%s' in topic list", name)
	}
	return nil
}
