package test

import (
	"bytes"
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

	"github.com/postie-labs/go-postie-lib/crypto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

const (
	DefaultTlsEnabled    = false
	DefaultHostAddr      = "localhost:8080"
	DefaultNumOfTopics   = 10
	DefaultNumOfUsers    = 100
	DefaultTopicLength   = 10
	DefaultRandomEnabled = false
	DefaultDebugEnabled  = false
)

var (
	tlsEnabled    bool
	hostAddr      string
	numOfTopics   int
	numOfUsers    int
	topicLength   int
	randomEnabled bool
	debugEnabled  bool
)

var Cmd = &cobra.Command{
	Use:   "test",
	Short: "run load test",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			conns []*grpc.ClientConn
			loop  = true
		)

		// init ctx, creds, wg
		ctx, cancel := context.WithCancel(context.Background())
		creds := grpc.WithTransportCredentials(insecure.NewCredentials())
		if tlsEnabled {
			creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
		}

		mainWg := sync.WaitGroup{}
		subWg := sync.WaitGroup{}

		// init sig channel
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// listen signals
		mainWg.Add(1)
		go func() {
			defer mainWg.Done()

			sig := <-sigCh
			fmt.Println("\r\ngot", sig.String())

			loop = false

			fmt.Printf("cancelling ctx ... ")
			cancel()
			fmt.Printf("done\n")

			subWg.Wait()

			for _, conn := range conns {
				fmt.Printf("closing grpc client ... ")
				err := conn.Close()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Printf("done\n")
			}
		}()

		// seed random
		rand.Seed(time.Now().UnixNano())

		for i := 0; i < numOfUsers && loop; i++ {
			userIndex := i
			if randomEnabled {
				userIndex = rand.Int()
			}
			nickname := fmt.Sprintf("alien-%d", userIndex)
			privKey, err := crypto.GenPrivKeyFromSeed([]byte(nickname))
			if err != nil {
				fmt.Println(err)
				continue
			}
			pubKeyBytes := privKey.PubKey().Bytes()

			// init grpc client
			conn, err := grpc.Dial(hostAddr, creds)
			if err != nil {
				fmt.Println(err)
				continue
			}
			conns = append(conns, conn)
			cli := pb.NewLakeClient(conn)

			for j := 0; j < numOfTopics && loop; j++ {
				topicIndex := j
				if randomEnabled {
					topicIndex = rand.Int()
				}
				topic := fmt.Sprintf("topic-%d", topicIndex)
				// execute goroutine (receiver)
				subWg.Add(1)
				go func() {
					defer subWg.Done()
					stream, err := cli.Recv(ctx, &pb.RecvReq{
						MsgBoxId:   topic,
						ConsumerId: nickname,
					})
					if err != nil {
						fmt.Println(err)
						return
					}
					for {
						select {
						case <-ctx.Done():
							return
						default:
							data, err := stream.Recv()
							if err == io.EOF || status.Code(err) > codes.OK {
								return
							}
							if err != nil {
								fmt.Println(err)
								return
							}
							msgCapsule := data.GetMsgCapsule()
							signature := msgCapsule.GetSignature()
							if bytes.Equal(signature.GetPubKey(), pubKeyBytes) {
								continue
							}
							msg := msgCapsule.GetMsg()
							metadata := msg.GetMetadata()
							nickname := "unknown"
							value, exist := metadata["nickname"]
							if exist {
								nickname = string(value)
							}
							if debugEnabled {
								fmt.Printf("[%s]%s:%s\n", topic, nickname, msg.GetData())
							}
						}
					}
				}()

				// execute goroutine (sender)
				subWg.Add(1)
				go func() {
					defer subWg.Done()
					ticker := time.NewTicker(1 * time.Second)
					defer ticker.Stop()
					sendClient, err := cli.Send(ctx)
					if err != nil {
						fmt.Println(err)
						return
					}
					for {
						select {
						case <-ctx.Done():
							res, err := sendClient.CloseAndRecv()
							if err != nil {
								fmt.Println(err)
								return
							}
							if !res.Ok {
								fmt.Println("failed to send msg")
								return
							}
							return
						case <-ticker.C:
							msg := &pb.Msg{
								Data: []byte(fmt.Sprintf("helloworld-%d", time.Now().UnixNano())),
								Metadata: map[string][]byte{
									"nickname": []byte(nickname),
								},
							}
							data, err := proto.Marshal(msg)
							if err != nil {
								fmt.Println(err)
								continue
							}
							sigBytes, err := privKey.Sign(data)
							if err != nil {
								fmt.Println(err)
								continue
							}

							err = sendClient.Send(&pb.SendReq{
								MsgBoxId: topic,
								MsgCapsule: &pb.MsgCapsule{
									Msg: msg,
									Signature: &pb.Signature{
										PubKey:   pubKeyBytes,
										SigBytes: sigBytes,
									},
								},
							})
							if err != nil {
								fmt.Println(err)
								return
							}
						}
					}
				}()
			}
			time.Sleep(1 * time.Millisecond)
		}
		fmt.Printf("successfully initiated clients: %d\n", numOfTopics*numOfUsers)
		mainWg.Wait()
		return nil
	},
}

func init() {
	Cmd.Flags().BoolVarP(&tlsEnabled, "tls", "t", DefaultTlsEnabled, "enable tls connection")
	Cmd.Flags().StringVar(&hostAddr, "host", DefaultHostAddr, "host addr")
	Cmd.Flags().IntVarP(&numOfTopics, "topics", "n", DefaultNumOfTopics, "number of topics")
	Cmd.Flags().IntVarP(&numOfUsers, "users", "u", DefaultNumOfUsers, "number of users")
	Cmd.Flags().IntVarP(&topicLength, "length", "l", DefaultTopicLength, "topic length")
	Cmd.Flags().BoolVarP(&randomEnabled, "random", "r", DefaultRandomEnabled, "enable random topic, nickname")
	Cmd.Flags().BoolVarP(&debugEnabled, "debug", "d", DefaultDebugEnabled, "enable debugging msgs")
}

func GenerateRandomString(length int) string {
	output := make([]byte, length)
	for i := 0; i < length; i++ {
		output[i] = byte(65 + rand.Intn(25))
	}
	return string(output)
}
