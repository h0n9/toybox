package agent

import (
	"context"
	"net"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/h0n9/toybox/msg-lake/lake"
	pb "github.com/h0n9/toybox/msg-lake/proto"
)

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "run msg lake agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		grpcServer := grpc.NewServer()
		lakeService, err := lake.NewLakeService(ctx)
		if err != nil {
			return err
		}

		pb.RegisterLakeServer(grpcServer, lakeService)

		listener, err := net.Listen("tcp", "0.0.0.0:8080")
		if err != nil {
			return err
		}

		err = grpcServer.Serve(listener)
		if err != nil {
			return err
		}

		return nil
	},
}
