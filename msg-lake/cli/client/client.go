package client

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	"github.com/postie-labs/go-postie-lib/crypto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

var (
	tlsEnabled bool
	hostAddr   string
	topicID    string
	nickname   string
)

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "run msg lake client (interactive)",
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
			if conn != nil {
				fmt.Printf("closing grpc client ... ")
				conn.Close()
				fmt.Printf("done\n")
			}
			fmt.Printf("cancelling ctx ... ")
			cancel()
			fmt.Printf("done\n")
		}()

		// init privKey
		privKey, err := crypto.GenPrivKeyFromSeed([]byte(nickname))
		if err != nil {
			return err
		}
		pubKeyBytes := privKey.PubKey().Bytes()

		// init grpc client
		creds := grpc.WithTransportCredentials(insecure.NewCredentials())
		if tlsEnabled {
			creds = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
		}
		conn, err = grpc.Dial(hostAddr, creds)
		if err != nil {
			return err
		}
		cli := pb.NewLakeClient(conn)

		stream, err := cli.Subscribe(ctx, &pb.SubscribeReq{
			TopicId: topicID,
		})
		if err != nil {
			return err
		}

		// block until recieve subscribe ack msg
		subRes, err := stream.Recv()
		if err != nil {
			return err
		}

		// check subscribe ack msg
		if subRes.GetType() != pb.SubscribeResType_SUBSCRIBE_RES_TYPE_ACK {
			return fmt.Errorf("failed to receive subscribe ack from agent")
		}
		if !subRes.GetOk() {
			return fmt.Errorf("failed to begin subscribing msgs")
		}

		// execute goroutine (receiver)
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				msg, err := stream.Recv()
				if err != nil {
					fmt.Println(err)
					sigCh <- syscall.SIGINT
					break
				}
				if msg.GetType() != pb.SubscribeResType_SUBSCRIBE_RES_TYPE_RELAY {
					continue
				}
				data := msg.GetData()
				if len(data) == 0 {
					continue
				}
				msgCapsule := pb.MsgCapsule{}
				err = proto.Unmarshal(data, &msgCapsule)
				if err != nil {
					fmt.Println(err)
					continue
				}

				signature := msgCapsule.GetSignature()
				if bytes.Equal(signature.GetPubKey(), pubKeyBytes) {
					continue
				}
				printOutput(true, msgCapsule.GetMsg())
				printInput(true)
			}
		}()

		// execute goroutine (sender)
		go func() {
			reader := bufio.NewReader(os.Stdin)
			for {
				printInput(false)
				input, err := reader.ReadString('\n')
				if err == io.EOF {
					return
				}
				if err != nil {
					fmt.Println(err)
					continue
				}
				input = strings.TrimSuffix(input, "\n")
				if input == "" {
					continue
				}
				msg := &pb.Msg{
					Data: []byte(input),
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
				msgCapsule := pb.MsgCapsule{
					Msg: msg,
					Signature: &pb.Signature{
						PubKey:   pubKeyBytes,
						SigBytes: sigBytes,
					},
				}
				data, err = proto.Marshal(&msgCapsule)
				if err != nil {
					fmt.Println(err)
					continue
				}

				pubRes, err := cli.Publish(ctx, &pb.PublishReq{
					TopicId: topicID,
					Data:    data,
				})
				if err == io.EOF {
					err := stream.CloseSend()
					if err != nil {
						fmt.Println(err)
					}
				}
				if err != nil {
					fmt.Println(err)
					continue
				}

				// check publish res
				if !pubRes.GetOk() {
					fmt.Println("failed to send message")
					continue
				}
			}
		}()

		wg.Wait()

		return nil
	},
}

func printInput(newline bool) {
	s := "💬 <%s> "
	if newline {
		s = "\r\n" + s
	}
	fmt.Printf(s, nickname)
}

func printOutput(newline bool, msg *pb.Msg) {
	s := "📩 <%s> %s"
	if newline {
		s = "\r\n" + s
	}
	nickname := "unknown"
	metadata := msg.GetMetadata()
	value, exist := metadata["nickname"]
	if exist {
		nickname = string(value)
	}
	fmt.Printf(s, nickname, msg.GetData())
}

func init() {
	r := rand.New(rand.NewSource(time.Now().Unix())).Int()

	Cmd.Flags().BoolVarP(&tlsEnabled, "tls", "t", false, "enable tls connection")
	Cmd.Flags().StringVar(&hostAddr, "host", "localhost:8080", "host addr")
	Cmd.Flags().StringVar(&topicID, "topic", "life is beautiful", "topic id")
	Cmd.Flags().StringVarP(&nickname, "nickname", "n", fmt.Sprintf("alien-%d", r), "consumer id")
}
