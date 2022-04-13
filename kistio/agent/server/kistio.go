package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	pb "github.com/h0n9/toybox/kistio/proto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type KistioServer struct {
	pb.UnimplementedKistioServer

	topicStates map[string]*TopicState
	node        *p2p.Node
	subWg       sync.WaitGroup
}

func NewKistioServer(node *p2p.Node) *KistioServer {
	return &KistioServer{
		topicStates: make(map[string]*TopicState),
		node:        node,
		subWg:       sync.WaitGroup{},
	}
}

func (server *KistioServer) Close() {
	server.subWg.Wait()
	keys := make([]string, 0, len(server.topicStates))
	for ts := range server.topicStates {
		keys = append(keys, ts)
	}
	fmt.Println("closing topicStates:", keys)
	for _, key := range keys {
		ts := server.topicStates[key]
		err := ts.Close()
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
	server.subWg.Add(1)
	defer server.subWg.Done()

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

	streamCtx := stream.Context()

	eh, err := ts.topicConsumer.EventHandler()
	if err != nil {
		fmt.Println(err)
	}
	defer eh.Cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			e, err := eh.NextPeerEvent(streamCtx)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(e, ts.topicConsumer.ListPeers())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := sub.Next(streamCtx)
			if err != nil {
				fmt.Println(err)
				return
			}
			err = stream.Send(&pb.SubscribeResponse{Data: msg.GetData()})
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}()

	wg.Wait()

	fmt.Println("end of subscription:", tpName)

	return nil
}
