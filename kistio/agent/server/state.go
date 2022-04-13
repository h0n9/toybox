package server

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	PostfixGossipTopic = "gossip-consumer"
)

type TopicState struct {
	topic         *pubsub.Topic
	topicConsumer *pubsub.Topic

	myID            peer.ID
	myPartition     int
	consumerPeerIDs []peer.ID
}

func NewTopicState(topic, topicConsumer *pubsub.Topic, myID peer.ID) (*TopicState, error) {
	return &TopicState{
		topic:         topic,
		topicConsumer: topicConsumer,

		myID:            myID,
		myPartition:     1,
		consumerPeerIDs: make([]peer.ID, 0),
	}, nil
}

func (ts *TopicState) Close() error {
	if ts.topicConsumer != nil {
		fmt.Println("close topic:", ts.topicConsumer.String())
		err := ts.topicConsumer.Close()
		if err != nil {
			return err
		}
	}
	fmt.Println("close topic:", ts.topic.String())
	err := ts.topic.Close()
	if err != nil {
		return err
	}
	return nil
}

func (ts *TopicState) Update() {
	peerIDs := ts.topicConsumer.ListPeers()

	if reflect.DeepEqual(ts.consumerPeerIDs, peerIDs) {
		return
	}

	peerIDs = append(peerIDs, ts.myID)
	sort.Slice(peerIDs, func(i int, j int) bool {
		return peerIDs[i] > peerIDs[j]
	})
	for i, peerID := range peerIDs {
		if peerID == ts.myID {
			ts.myPartition = i + 1
			break
		}
	}

	ts.consumerPeerIDs = peerIDs[:len(peerIDs)-1]

	fmt.Println("updated myPartition:", ts.myPartition)
}
