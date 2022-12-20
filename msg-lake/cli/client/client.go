package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/h0n9/toybox/msg-lake/proto"
)

var (
	hostAddr               string
	msgBoxID               string
	producerID, consumerID string
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
		conn, err := grpc.Dial(hostAddr, grpcOpts...)
		if err != nil {
			return err
		}
		cli := proto.NewLakeClient(conn)

		// execute goroutine (receiver)
		wg.Add(1)
		go func() {
			defer wg.Done()

			stream, err := cli.Recv(ctx, &proto.RecvReq{
				MsgBoxId:   msgBoxID,
				ConsumerId: consumerID,
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
					break
				}

				msg := data.GetMsg()
				if msg.GetFrom().GetAddress() == producerID {
					continue
				}
				fmt.Printf("\r\n📩 <%s> %s\r\n", msg.GetFrom().GetAddress(), msg.GetData().GetData())
			}

		}()

		// execute goroutine (sender)
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				fmt.Printf("\r\n️️💬 <%s> ", producerID)
				input, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
					continue
				}
				input = strings.TrimSuffix(input, "\n")
				if input == "" {
					continue
				}
				res, err := cli.Send(ctx, &proto.SendReq{
					MsgBoxId: msgBoxID,
					Msg: &proto.Msg{
						From: &proto.Address{
							Address: producerID,
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
		}()

		wg.Wait()

		return nil
	},
}

func init() {
	r := rand.New(rand.NewSource(time.Now().Unix())).Int()

	Cmd.Flags().StringVar(&hostAddr, "host", "localhost:8080", "host addr")
	Cmd.Flags().StringVarP(&msgBoxID, "box", "b", "test", "msg box id")
	Cmd.Flags().StringVarP(&producerID, "producer", "p", fmt.Sprintf("test-producer-%d", r), "producer id")
	Cmd.Flags().StringVarP(&consumerID, "consumer", "c", fmt.Sprintf("test-consumer-%d", r), "consumer id")
}
