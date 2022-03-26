package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	"github.com/h0n9/toybox/kistio/agent/server"
	"github.com/h0n9/toybox/kistio/agent/util"
	pb "github.com/h0n9/toybox/kistio/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// prepare
	ctx, cancel := context.WithCancel(context.Background())
	cfg := util.NewConfig()
	cfg.ParseFlags()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// init node
	node, err := p2p.NewNode(ctx, cfg)
	if err != nil {
		panic(err)
	}

	err = node.Bootstrap(cfg.NodeBootstraps)
	if err != nil {
		panic(err)
	}

	go node.Discover(cfg.RendezVous)

	fmt.Println(node.Info())

	// init server(handler)
	kistioSrv := server.NewKistioServer(node)
	healthCheckerSrv := server.NewHealthChecker()

	// init grpcServer
	opts := []grpc.ServerOption{}
	grpcSrv := grpc.NewServer(opts...)

	// register server(handler) to grpcServer
	pb.RegisterKistioServer(grpcSrv, kistioSrv)
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthCheckerSrv)

	// start grpcServer
	listener, err := net.Listen("tcp", cfg.GrpcListen)
	if err != nil {
		panic(err)
	}

	go func() {
		sig := <-sigs // block until signal
		fmt.Printf("\nRECEIVED SIGNAL: %s\n", sig)

		cancel()               // cancel context
		grpcSrv.GracefulStop() // gracefully stop grpcServer
		kistioSrv.Close()      // close kistioServer
		node.Close()           // close node
		if err != nil {
			fmt.Println(err)
		}
	}()

	err = grpcSrv.Serve(listener)
	if err != nil {
		panic(err)
	}
}
