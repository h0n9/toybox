package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/h0n9/toybox/kistio/proto"
)

const (
	DefaultAgentHost = "127.0.0.1"
	DefaultAgentPort = "7788"
	DefaultTopicPub  = "balance"
	DefaultTopicSub  = "account"
	DefaultInterval  = int64(1000) // millisecond
)

type Msg struct {
	Sender   string `json:"sender"`
	Data     []byte `json:"data"`
	Metadata []byte `json:"metadata"`
}

func main() {
	// init flags
	agentHost := flag.String("agent-host", DefaultAgentHost, "agent grpc host")
	agentPort := flag.String("agent-port", DefaultAgentPort, "agent grpc port")
	topicPub := flag.String("topic-pub", DefaultTopicPub, "topic for publish")
	topicSub := flag.String("topic-sub", DefaultTopicSub, "topic for subscribe")
	interval := flag.Int64("interval", DefaultInterval, "interval for publish")
	flag.Parse()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// init grpc.Dial
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	conn, err := grpc.Dial(net.JoinHostPort(*agentHost, *agentPort), opts...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// init client
	cli := pb.NewKistioClient(conn)

	// init context
	ctx, cancel := context.WithCancel(context.Background())

	// handle signals
	go func() {
		sig := <-sigs
		fmt.Printf("\nRECEIVED SIGNAL: %s\n", sig)

		cancel()
	}()

	// init waitGroup
	wg := sync.WaitGroup{}

	// init goroutine for publish
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(time.Duration(*interval) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("stop sending msgs")
				return
			case <-ticker.C:
				// init msg to send
				msg := Msg{
					Sender:   *topicSub,
					Data:     []byte("I'd like to buy an apple"),
					Metadata: []byte(time.Now().String()),
				}
				data, err := json.Marshal(msg)
				if err != nil {
					fmt.Println(err)
					continue
				}

				res, err := cli.Publish(ctx, &pb.PublishRequest{
					Topic: *topicPub,
					Data:  data,
				})
				if err != nil {
					fmt.Println(err)
					continue
				}

				if !res.Ok {
					fmt.Println("failed to publish msg")
				}

				fmt.Printf("client-pub: %s\n", data)
			}
		}
	}()

	// init goroutine for subscribe
	wg.Add(1)
	go func() {
		defer wg.Done()
		stream, err := cli.Subscribe(ctx, &pb.SubscribeRequest{Topic: *topicSub})
		if err != nil {
			panic(err)
		}
		defer stream.CloseSend()
		for {
			msg, err := stream.Recv()
			if err == io.EOF || status.Code(err) == codes.Canceled {
				fmt.Println("stop stream receiving msgs")
				break
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("client-sub: %s\n", msg.GetData())
		}
	}()

	wg.Wait()
}
