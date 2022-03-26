package server

import (
	"context"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/grpc/health/grpc_health_v1"

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
	fmt.Printf("server-pub: %s\n", data)
	return &pb.PublishResponse{Ok: true}, nil
}

func (server *KistioServer) Subscribe(req *pb.SubscribeRequest, stream pb.Kistio_SubscribeServer) error {
	// check
	tpName := req.GetTopic()

	// execute
	tp, err := server.getTopic(tpName)
	if err != nil {
		return err
	}
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
		fmt.Printf("server-sub: %s\n", data)
		err = stream.Send(&pb.SubscribeResponse{Data: data})
		if err != nil {
			return err
		}
	}
}

type HealthChecker struct{}

func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}

func (server *HealthChecker) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (server *HealthChecker) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	return stream.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}
