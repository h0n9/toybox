package server

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/h0n9/toybox/msg-lake/lake"
	"github.com/h0n9/toybox/msg-lake/proto"
)

var (
	listenAddr string
)

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "run msg lake server",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			grpcServer *grpc.Server
			lakeServer *lake.LakeServer
		)

		// init wg
		wg := sync.WaitGroup{}

		// init listener
		listener, err := net.Listen("tcp", listenAddr)
		if err != nil {
			return err
		}

		// init sig channel
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// listen signals
		wg.Add(1)
		go func() {
			defer wg.Done()

			sig := <-sigCh

			fmt.Println("\r\ngot", sig.String())
			if grpcServer != nil {
				fmt.Printf("closing grpc server ... ")
				grpcServer.GracefulStop()
				fmt.Printf("done\n")
			}
			if lakeServer != nil {
				fmt.Printf("closing lake server ... ")
				lakeServer.Close()
				fmt.Printf("done\n")
			}
			if listener != nil {
				fmt.Printf("closing listener ... ")
				listener.Close()
				fmt.Printf("done\n")
			}
		}()

		// init grpc, lake servers and register lakeServer to grpcServer
		grpcServer = grpc.NewServer()
		lakeServer = lake.NewLakeServer()
		proto.RegisterLakeServer(grpcServer, lakeServer)

		wg.Add(1)
		go func() {
			defer wg.Done()

			fmt.Printf("listening on %s\n", listenAddr)

			err := grpcServer.Serve(listener)
			if err != nil {
				fmt.Println(err)
			}
		}()

		wg.Wait()

		return nil
	},
}

func init() {
	Cmd.Flags().StringVarP(&listenAddr, "listen", "l", "0.0.0.0:8080", "listening addr")
}
