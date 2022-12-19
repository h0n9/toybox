package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/h0n9/toybox/msg-lake/proto"
)

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "runs msg lake client (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			conn *grpc.ClientConn
		)
		// init wg
		wg := sync.WaitGroup{}

		// init sig channel
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// init ctx with cancel
		ctx, cancel := context.WithCancel(context.Background())

		// listen signals
		wg.Add(1)
		go func() {
			defer wg.Done()

			sig := <-sigCh
			fmt.Println("\r\ngot", sig.String())

			fmt.Printf("cancelling ctx ... ")
			cancel()
			fmt.Printf("done\n")

			if conn != nil {
				fmt.Printf("closing grpc client ... ")
				conn.Close()
				fmt.Printf("done\n")
			}
		}()

		// init grpc client
		grpcOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		conn, err := grpc.Dial("localhost:8080", grpcOpts...)
		if err != nil {
			return err
		}
		cli := proto.NewLakeClient(conn)

		// execute goroutine (receiver)
		wg.Add(1)
		go func() {
			defer wg.Done()

			stream, err := cli.Recv(ctx, &proto.RecvReq{
				MsgBoxId:   "test",
				ConsumerId: "test-consumer-0",
			})
			if err != nil {
				fmt.Println(err)
				return
			}

			for {
				data, err := stream.Recv()
				if err == io.EOF || status.Code(err) == codes.Canceled {
					fmt.Println("stop receiving msgs")
					break
				}
				if err != nil {
					fmt.Println(err)
					continue
				}

				msg := data.GetMsg()
				fmt.Printf("%s> %s", msg.GetFrom(), msg.GetData())
			}

		}()

		// execute goroutine (sender)
		wg.Add(1)
		go func() {
			defer wg.Done()
			var input string
			for loop := true; loop; {
				select {
				case <-ctx.Done():
					loop = false
				default:
					fmt.Printf("\r\n🎙️> ")
					fmt.Scanln(&input)
					res, err := cli.Send(ctx, &proto.SendReq{
						MsgBoxId: "test",
						Msg: &proto.Msg{
							From: &proto.Address{
								Address: "test-producer-0",
							},
							Data: &proto.Data{
								Data: []byte(input),
							},
						},
					})
					if err != nil {
						fmt.Println(err)
						continue
					}

					if !res.Ok {
						fmt.Println("failed to send msg")
					}
				}
			}
		}()

		wg.Wait()

		return nil
	},
}
