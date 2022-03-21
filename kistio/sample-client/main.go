package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h0n9/toybox/kistio/proto"
)

const (
	DefaultAgentHost = "127.0.0.1"
	DefaultAgentPort = "7788"
	DefaultTopicPub  = "balance"
	DefaultTopicSub  = "account"
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
	flag.Parse()

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
	ctx := context.Background()
	defer ctx.Done()

	// init waitGroup
	wg := sync.WaitGroup{}

	// init goroutine for publish
	wg.Add(1)
	go func() {
		for {
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

			time.Sleep(1 * time.Second)
		}
	}()

	// init goroutine for subscribe
	wg.Add(1)
	go func() {
		stream, err := cli.Subscribe(ctx, &pb.SubscribeRequest{Topic: *topicSub})
		if err != nil {
			panic(err)
		}
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s\n", msg.GetData())
		}
	}()

	wg.Wait()
}
