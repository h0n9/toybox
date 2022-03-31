package server

import (
	"context"
	"fmt"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	pb "github.com/h0n9/toybox/kistio/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type KistioServer struct {
	pb.UnimplementedKistioServer

	topicStates map[string]*TopicState
	node        *p2p.Node
}

func NewKistioServer(node *p2p.Node) *KistioServer {
	return &KistioServer{
		node:        node,
		topicStates: make(map[string]*TopicState),
	}
}

func (server *KistioServer) Close() {
	keys := make([]string, 0, len(server.topicStates))
	for topic := range server.topicStates {
		keys = append(keys, topic)
	}
	for _, key := range keys {
		tp := server.topicStates[key]
		err := tp.Close()
		if err != nil {
			fmt.Println(err)
		}
		delete(server.topicStates, key)
	}
}

func (server *KistioServer) getTopicState(name string, consume bool) (*TopicState, error) {
	ts, exist := server.topicStates[name]
	if !exist {
		topic, err := server.node.Join(name)
		if err != nil {
			return nil, err
		}
		var gossipTopic *pubsub.Topic = nil
		if consume {
			gossipTopic, err = server.node.Join(fmt.Sprintf("%s-%s", topic.String(), PostfixGossipTopic))
			if err != nil {
				return nil, err
			}
		}
		tmp, err := NewTopicState(topic, gossipTopic, server.node.GetHostID())
		if err != nil {
			return nil, err
		}
		server.topicStates[name] = tmp
		ts = tmp
	}
	return ts, nil
}

func (server *KistioServer) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// check
	tpName := req.GetTopic()
	data := req.GetData()

	// execute
	ts, err := server.getTopicState(tpName, false)
	if err != nil {
		return nil, err
	}
	err = ts.topic.Publish(ctx, data)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("server-pub: %s\n", data)
	return &pb.PublishResponse{Ok: true}, nil
}

func (server *KistioServer) Subscribe(req *pb.SubscribeRequest, stream pb.Kistio_SubscribeServer) error {
	// check
	tpName := req.GetTopic()

	// execute
	ts, err := server.getTopicState(tpName, true)
	if err != nil {
		return err
	}
	defer ts.Close()
	sub, err := ts.topic.Subscribe()
	if err != nil {
		return err
	}
	defer sub.Cancel()
	subConsumer, err := ts.topicConsumer.Subscribe()
	if err != nil {
		return err
	}
	defer subConsumer.Cancel()

	ctx := stream.Context()

	//ts.topicConsumer.EventHandler(func(t *pubsub.TopicEventHandler) error {
	//	for {
	//		peerEvent, err := t.NextPeerEvent(ctx)
	//		if err != nil {
	//			return err
	//		}
	//		fmt.Println(peerEvent)
	//	}
	//})

	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			return err
		}
		data := msg.GetData()
		// fmt.Printf("server-sub: %s\n", data)
		err = stream.Send(&pb.SubscribeResponse{Data: data})
		if err != nil {
			return err
		}
	}
}
