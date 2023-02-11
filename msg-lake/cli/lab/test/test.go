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

	"github.com/postie-labs/go-postie-lib/crypto"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

const (
	DefaultTlsEnabled    = false
	DefaultHostAddr      = "localhost:8080"
	DefaultNumOfUsers    = 100
	DefaultTopicLength   = 10
	DefaultRandomEnabled = false
	DefaultDebugEnabled  = false
)

var (
	tlsEnabled    bool
	hostAddr      string
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

		msgBoxID := "topic-0"
		if randomEnabled {
			msgBoxID = fmt.Sprintf("topic-%d", rand.Int())
		}

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

			// send
			subWg.Add(1)
			go func() {
				defer subWg.Done()
				ticker := time.NewTicker(1 * time.Second)
				defer ticker.Stop()
				sender, err := cli.Send(ctx)
				if err != nil {
					fmt.Println(err)
					return
				}
				msg := &pb.Msg{
					Metadata: map[string][]byte{
						"nickname": []byte(nickname),
					},
				}
				req := &pb.SendReq{
					MsgBoxId: msgBoxID,
					MsgCapsule: &pb.MsgCapsule{
						Signature: &pb.Signature{
							PubKey: pubKeyBytes,
						},
					},
				}
				for {
					select {
					case <-ctx.Done():
						err = sender.CloseSend()
						if err != nil {
							fmt.Println(err)
						}
						return
					case <-ticker.C:
						msg.Data = []byte(fmt.Sprintf("helloworld-%d", time.Now().UnixNano()))
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
						req.MsgCapsule.Msg = msg
						req.MsgCapsule.Signature.SigBytes = sigBytes
						err = sender.Send(req)
						if err == io.EOF {
							return
						}
						if err != nil {
							fmt.Println(err)
							continue
						}
					}
				}
			}()

			// recv
			subWg.Add(1)
			go func() {
				subWg.Done()
				receiver, err := cli.Recv(ctx, &pb.RecvReq{
					MsgBoxId:   msgBoxID,
					ConsumerId: nickname,
				})
				if err != nil {
					fmt.Println(err)
					return
				}
				for {
					select {
					case <-ctx.Done():
						err = receiver.CloseSend()
						if err != nil {
							fmt.Println(err)
						}
						return
					default:
						data, err := receiver.Recv()
						if err == io.EOF {
							return
						}
						data.GetMsgCapsule()
					}
				}
			}()
		}

		fmt.Printf("successfully initiated clients: %d, msgBoxID: %s\n", numOfUsers, msgBoxID)
		mainWg.Wait()
		return nil
	},
}

func init() {
	Cmd.Flags().BoolVarP(&tlsEnabled, "tls", "t", DefaultTlsEnabled, "enable tls connection")
	Cmd.Flags().StringVar(&hostAddr, "host", DefaultHostAddr, "host addr")
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
