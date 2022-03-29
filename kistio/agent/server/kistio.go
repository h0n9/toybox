package server

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	pb "github.com/h0n9/toybox/kistio/proto"
)

type KistioServer struct {
	pb.UnimplementedKistioServer

	topics map[string]*pubsub.Topic
	node   *p2p.Node
}

func NewKistioServer(node *p2p.Node) *KistioServer {
	return &KistioServer{
		node:   node,
		topics: make(map[string]*pubsub.Topic),
	}
}

func (server *KistioServer) Close() {
	keys := make([]string, 0, len(server.topics))
	for topic := range server.topics {
		keys = append(keys, topic)
	}
	for _, key := range keys {
		tp := server.topics[key]
		fmt.Println("close topic:", key)
		err := tp.Close()
		if err != nil {
			fmt.Println(err)
		}
		delete(server.topics, key)
	}
}

func (server *KistioServer) getTopic(name string) (*pubsub.Topic, error) {
	topic, exist := server.topics[name]
	if !exist {
		tmp, err := server.node.Join(name)
		if err != nil {
			return nil, err
		}
		server.topics[name] = tmp
		topic = tmp
	}
	return topic, nil
}

func (server *KistioServer) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// check
	tpName := req.GetTopic()
	data := req.GetData()

	// execute
	tp, err := server.getTopic(tpName)
	if err != nil {
		return nil, err
	}
	err = tp.Publish(ctx, data)
	if err != nil {
		return nil, err
	}
	fmt.Println("# of topic peers:", len(tp.ListPeers()))
	// fmt.Printf("server-pub: %s\n", data)
	return &pb.PublishResponse{Ok: true}, nil
}

func (server *KistioServer) Subscribe(req *pb.SubscribeRequest, stream pb.Kistio_SubscribeServer) error {
	// check
	tpName := req.GetTopic()

	// gossip consumer
	// go server.gossipConsumer(tpName)

	// execute
	tp, err := server.getTopic(tpName)
	if err != nil {
		return err
	}
	defer tp.Close()
	sub, err := tp.Subscribe()
	if err != nil {
		return err
	}
	defer sub.Cancel()
	ctx := stream.Context()
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

func (server *KistioServer) gossipConsumer(tpName string) {
	tpGCName := fmt.Sprintf("%s-%s-%s", tpName, "gossip", "consumer")
	tpGC, err := server.getTopic(tpGCName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tpGC.Close()
}
