package main

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/h0n9/toybox/msg-lake/lake"
	pb "github.com/h0n9/toybox/msg-lake/proto"
)

func main() {
	ctx := context.Background()

	grpcServer := grpc.NewServer()
	lakeService, err := lake.NewLakeService(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	pb.RegisterLakeServer(grpcServer, lakeService)

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		fmt.Println(err)
		return
	}
}
