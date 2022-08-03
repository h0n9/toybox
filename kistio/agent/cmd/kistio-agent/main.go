package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/h0n9/toybox/kistio/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	"github.com/h0n9/toybox/kistio/agent/server"
	"github.com/h0n9/toybox/kistio/agent/util"
)

func main() {
	// prepare
	ctx, cancel := context.WithCancel(context.Background())
	cfg := util.NewConfig()
	cfg.ParseFlags()

	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	done := make(chan bool, 1)
	defer close(done)

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
	grpcSrv := grpc.NewServer()

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

		grpcSrv.GracefulStop() // gracefully stop grpcServer
		kistioSrv.Close()      // close kistioServer
		err = node.Close()     // close node
		if err != nil {
			fmt.Println(err)
		}
		cancel() // cancel context
		fmt.Println("-------------- closed all --------------")
		done <- true
	}()

	err = grpcSrv.Serve(listener)
	if err != nil {
		panic(err)
	}

	<-done

	fmt.Println("-------------- end of main --------------")
}
