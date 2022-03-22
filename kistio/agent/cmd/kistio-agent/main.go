package main

import (
	"context"
	"fmt"
	"net"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	"github.com/h0n9/toybox/kistio/agent/server"
	"github.com/h0n9/toybox/kistio/agent/util"
	pb "github.com/h0n9/toybox/kistio/proto"
	"google.golang.org/grpc"
)

func main() {
	// prepare
	ctx := context.Background()
	cfg := util.NewConfig()
	cfg.ParseFlags()

	// init node
	node, err := p2p.NewNode(ctx, cfg)
	if err != nil {
		panic(err)
	}

	err = node.Bootstrap(cfg.NodeBootstraps)
	if err != nil {
		panic(err)
	}

	go node.DiscoverPeers()

	fmt.Println(node.Info())

	// init server(handler)
	srv := server.NewKistioServer(node)

	// init grpcServer
	opts := []grpc.ServerOption{}
	grpcSrv := grpc.NewServer(opts...)

	// register server(handler) to grpcServer
	pb.RegisterKistioServer(grpcSrv, srv)

	// start grpcServer
	listener, err := net.Listen("tcp", cfg.GrpcListen)
	if err != nil {
		panic(err)
	}
	err = grpcSrv.Serve(listener)
	if err != nil {
		panic(err)
	}

	// tp, err := node.Join("test")
	// if err != nil {
	// 	panic(err)
	// }

	// wg := sync.WaitGroup{}

	// // publish
	// wg.Add(1)
	// go func() {
	// 	for {
	// 		// publish simple data for test
	// 		err = tp.Publish(ctx, []byte(fmt.Sprintf("%s - hello world", time.Now().String())))
	// 		time.Sleep(1 * time.Second)
	// 	}
	// }()

	// // subscribe
	// wg.Add(1)
	// go func() {
	// 	sub, err := tp.Subscribe()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	for {
	// 		msg, err := sub.Next(ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			continue
	// 		}
	// 		if msg.GetFrom() == node.GetHostID() {
	// 			continue
	// 		}
	// 		fmt.Printf("[%s] %s\n", msg.GetFrom().Pretty(), msg.GetData())
	// 	}
	// }()

	// wg.Wait()
}
