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

var (
	tlsEnabled bool
	hostAddr   string
	msgBoxID   string
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

		// execute goroutine (receiver)
		wg.Add(1)
		go func() {
			defer wg.Done()

			stream, err := cli.Recv(ctx, &pb.RecvReq{
				MsgBoxId:   msgBoxID,
				ConsumerId: nickname,
			})
			if err != nil {
				fmt.Println(err)
				return
			}

			for {
				data, err := stream.Recv()
				if err == io.EOF || status.Code(err) > codes.OK {
					fmt.Println("stop receiving msgs")
					break
				}
				if err != nil {
					fmt.Println(err)
					sigCh <- syscall.SIGINT
					break
				}

				msgCapsule := data.GetMsgCapsule()
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
			sendClient, err := cli.Send(ctx)
			if err != nil {
				fmt.Println(err)
				sigCh <- syscall.SIGINT
				return
			}
			for {
				printInput(false)
				input, err := reader.ReadString('\n')
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

				err = sendClient.Send(&pb.SendReq{
					MsgBoxId: msgBoxID,
					MsgCapsule: &pb.MsgCapsule{
						Msg: msg,
						Signature: &pb.Signature{
							PubKey:   pubKeyBytes,
							SigBytes: sigBytes,
						},
					},
				})
				if err == io.EOF {
					res, err := sendClient.CloseAndRecv()
					if err != nil {
						fmt.Println(err)
					}
					if !res.GetOk() {
						fmt.Println("something went wrong on closing send client")
					}
					return
				}
				if err != nil {
					fmt.Println(err)
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
	Cmd.Flags().StringVarP(&msgBoxID, "box", "b", "life is beautiful", "msg box id")
	Cmd.Flags().StringVarP(&nickname, "nickname", "n", fmt.Sprintf("alien-%d", r), "consumer id")
}
