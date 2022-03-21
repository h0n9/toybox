package server

import (
	"context"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	pb "github.com/h0n9/toybox/kistio/proto"
)

type KistioServer struct {
	pb.UnimplementedKistioServer

	node *p2p.Node
}

func NewKistioServer(node *p2p.Node) *KistioServer {
	return &KistioServer{node: node}
}

func (server *KistioServer) Publish(ctx context.Context, req *pb.PublishRequest) (*pb.PublishResponse, error) {
	// check
	tpStr := req.GetTopic()
	data := req.GetData()

	// execute
	tp, err := server.node.Join(tpStr)
	if err != nil {
		return nil, err
	}
	err = tp.Publish(ctx, data)
	if err != nil {
		return nil, err
	}
	return &pb.PublishResponse{Ok: true}, nil
}

func (server *KistioServer) Subscribe(req *pb.SubscribeRequest, stream pb.Kistio_SubscribeServer) error {
	// check
	tpStr := req.GetTopic()

	// execute
	tp, err := server.node.Join(tpStr)
	if err != nil {
		return err
	}
	sub, err := tp.Subscribe()
	if err != nil {
		return err
	}
	ctx := stream.Context()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			return err
		}
		data := msg.GetData()
		stream.Send(&pb.SubscribeResponse{Data: data})
	}
}
