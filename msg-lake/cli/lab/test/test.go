package test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/h0n9/toybox/msg-lake/proto"
)

const (
	DefaultTlsEnabled  = false
	DefaultHostAddr    = "localhost:8080"
	DefaultNumOfTopics = 10
	DefaultNumOfUsers  = 100
	DefaultTopicLength = 10
)

var (
	tlsEnabled  bool
	hostAddr    string
	numOfTopics int
	numOfUsers  int
	topicLength int
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "run load test",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			conns []*grpc.ClientConn
		)

		// init ctx, creds, wg
		ctx, cancel := context.WithCancel(context.Background())
		creds := grpc.WithTransportCredentials(insecure.NewCredentials())
		if tlsEnabled {
			creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
		}
		wg := sync.WaitGroup{}

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

			fmt.Printf("cancelling ctx ... ")
			cancel()
			fmt.Printf("done\n")

			for _, conn := range conns {
				fmt.Printf("closing grpc client ... ")
				conn.Close()
				fmt.Printf("done\n")
			}
		}()

		// seed random
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < numOfTopics; i++ {
			topic := GenerateRandomString(topicLength)
			for j := 0; j < numOfUsers; j++ {
				// generate random consumer id
				nickname := fmt.Sprintf("alien-%d", rand.Int())

				// init grpc client
				conn, err := grpc.Dial(hostAddr, creds)
				if err != nil {
					fmt.Println(err)
					continue
				}
				conns = append(conns, conn)
				cli := proto.NewLakeClient(conn)

				// execute goroutine (receiver)
				wg.Add(1)
				go func() {
					defer wg.Done()
					stream, err := cli.Recv(ctx, &proto.RecvReq{
						MsgBoxId:   topic,
						ConsumerId: nickname,
					})
					if err != nil {
						fmt.Println(err)
						return
					}
					for {
						data, err := stream.Recv()
						if err == io.EOF || status.Code(err) == codes.Canceled {
							fmt.Println("stop receiving msgs")
							return
						}
						if err != nil {
							fmt.Println(err)
							sigCh <- syscall.SIGINT
							return
						}

						msg := data.GetMsg()
						if msg.GetFrom().GetAddress() == nickname {
							continue
						}
						fmt.Printf("[%s]%s:%s\n", topic, msg.GetFrom().GetAddress(), msg.GetData().GetData())
					}
				}()

				// execute goroutine (sender)
				wg.Add(1)
				go func() {
					defer wg.Done()
					ticker := time.NewTicker(1 * time.Second)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							data := []byte(fmt.Sprintf("helloworld-%d", time.Now().UnixNano()))
							res, err := cli.Send(ctx, &proto.SendReq{
								MsgBoxId: topic,
								Msg: &proto.Msg{
									From: &proto.Address{
										Address: nickname,
									},
									Data: &proto.Data{
										Data: data,
									},
								},
							})
							if err != nil {
								fmt.Println(err)
								return
							}
							if !res.Ok {
								fmt.Println("failed to send msg")
							}
						}
					}
				}()
			}
		}
		wg.Wait()
		return nil
	},
}

func init() {
	Cmd.Flags().BoolVarP(&tlsEnabled, "tls", "t", DefaultTlsEnabled, "enable tls connection")
	Cmd.Flags().StringVar(&hostAddr, "host", DefaultHostAddr, "host addr")
	Cmd.Flags().IntVarP(&numOfTopics, "topics", "n", DefaultNumOfTopics, "number of topics")
	Cmd.Flags().IntVarP(&numOfUsers, "users", "u", DefaultNumOfUsers, "number of users")
	Cmd.Flags().IntVarP(&topicLength, "length", "l", DefaultTopicLength, "topic length")
}

func GenerateRandomString(length int) string {
	output := make([]byte, length)
	for i := 0; i < length; i++ {
		output[i] = byte(65 + rand.Intn(25))
	}
	return string(output)
}
