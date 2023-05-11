package lake

import (
	"context"
	"fmt"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type TopicManager struct {
	ctx        context.Context
	topics     map[string]*pubsub.Topic // topic_id:topic
	ps         *pubsub.PubSub
	subMsgChan chan pb.PubSubRes
	subWg      sync.WaitGroup
}

func NewTopicManager(ctx context.Context, ps *pubsub.PubSub, subMsgChan chan pb.PubSubRes) *TopicManager {
	return &TopicManager{
		ctx:        ctx,
		topics:     make(map[string]*pubsub.Topic),
		ps:         ps,
		subMsgChan: subMsgChan,
	}
}

func (tm *TopicManager) Close() {
	for _, topic := range tm.topics {
		// TODO: error handling
		topic.Close()
	}
	tm.subWg.Wait()
}

func (tm *TopicManager) Join(id string) (*pubsub.Topic, error) {
	topic := tm.getTopic(id)
	if topic == nil {
		newTopic, err := tm.ps.Join(id)
		if err != nil {
			return nil, err
		}
		topic = tm.setTopic(id, newTopic)

		tm.subWg.Add(1)
		go tm.subscribe(topic)
	}
	return topic, nil
}

func (tm *TopicManager) subscribe(topic *pubsub.Topic) error {
	defer tm.subWg.Done()

	sub, err := topic.Subscribe()
	if err != nil {
		return err
	}
	for {
		fmt.Println(topic.ListPeers())
		msgRaw, err := sub.Next(tm.ctx)
		if err != nil {
			return err
		}
		msgCapsule := pb.MsgCapsule{}
		err = proto.Unmarshal(msgRaw.GetData(), &msgCapsule)
		if err != nil {
			fmt.Println(err)
			continue
		}
		tm.subMsgChan <- pb.PubSubRes{
			Type:       pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
			TopicId:    msgRaw.GetTopic(),
			MsgCapsule: &msgCapsule,
		}
	}
}

func (tm *TopicManager) getTopic(id string) *pubsub.Topic {
	topic, exist := tm.topics[id]
	if !exist {
		return nil
	}
	return topic
}

func (tm *TopicManager) setTopic(id string, topic *pubsub.Topic) *pubsub.Topic {
	tm.topics[id] = topic
	return topic
}
