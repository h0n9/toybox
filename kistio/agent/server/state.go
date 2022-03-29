package server

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	PostfixGossipTopic = "gossip"
)

type TopicState struct {
	gossipTopic string
	consumers   []peer.AddrInfo
}

func NewTopicState(topic string) *TopicState {
	return &TopicState{
		gossipTopic: fmt.Sprintf("%s-%s", topic, PostfixGossipTopic),
		consumers:   make([]peer.AddrInfo, 0),
	}
}

func (ts *TopicState) Append(consumer peer.AddrInfo) error {
	return nil
}
