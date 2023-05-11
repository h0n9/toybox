package lake

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"google.golang.org/protobuf/proto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/relayer"
)

type LakeService struct {
	pb.UnimplementedLakeServer

	ctx     context.Context
	relayer *relayer.Relayer
}

func NewLakeService(ctx context.Context) (*LakeService, error) {
	relayer, err := relayer.NewRelayer(ctx, "0.0.0.0", 7733)
	if err != nil {
		return nil, err
	}
	go relayer.DiscoverPeers()
	return &LakeService{
		ctx:     ctx,
		relayer: relayer,
	}, nil
}

func (lakeService *LakeService) Close() error {
	var err error
	if lakeService.relayer != nil {
		err = lakeService.relayer.Close()
	}
	return err
}

func (lakeService *LakeService) PubSub(stream pb.Lake_PubSubServer) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	var (
		topics map[string]*pubsub.Topic = make(map[string]*pubsub.Topic)
		resCh  chan *pb.PubSubRes       = make(chan *pb.PubSubRes)
	)

	port := rand.Intn(8080-1000+1) + 1000
	h, err := lakeService.relayer.NewSubHost(port)
	if err != nil {
		return err
	}
	ps, err := pubsub.NewGossipSub(stream.Context(), h)
	if err != nil {
		return err
	}
	defer func() {
		for _, topic := range topics {
			err := topic.Close()
			if err != nil {
				fmt.Println(err)
			}
		}
		wg.Wait()
		err = h.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// req stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			fmt.Println("waiting")
			pubSubReq, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(pubSubReq)

			switch pubSubReq.Type {
			case pb.PubSubReqType_PUB_SUB_REQ_TYPE_PUBLISH:
				topic, exist := topics[pubSubReq.TopicId]
				if !exist {
					topic, err := ps.Join(pubSubReq.TopicId)
					if err != nil {
						resCh <- &pb.PubSubRes{
							Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
							Ok:   false,
						}
						continue
					}
					topics[pubSubReq.TopicId] = topic
				}
				data, err := proto.Marshal(pubSubReq.GetMsgCapsule())
				if err != nil {
					resCh <- &pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				err = topic.Publish(stream.Context(), data)
				if err != nil {
					resCh <- &pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				resCh <- &pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
					Ok:   true,
				}
			case pb.PubSubReqType_PUB_SUB_REQ_TYPE_SUBSCRIBE:
				topic, exist := topics[pubSubReq.TopicId]
				if !exist {
					topic, err := ps.Join(pubSubReq.TopicId)
					if err != nil {
						resCh <- &pb.PubSubRes{
							Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
							Ok:   false,
						}
						continue
					}
					topics[pubSubReq.TopicId] = topic
				}
				wg.Add(1)
				go func() {
					defer wg.Done()

					sub, err := topic.Subscribe()
					if err != nil {
						fmt.Println(err)
						return
					}

					for {
						msgRaw, err := sub.Next(stream.Context())
						if err != nil {
							fmt.Println(err)
							continue
						}
						msgCapsule := pb.MsgCapsule{}
						err = proto.Unmarshal(msgRaw.GetData(), &msgCapsule)
						if err != nil {
							fmt.Println(err)
							continue
						}
						resCh <- &pb.PubSubRes{
							Type:       pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
							TopicId:    msgRaw.GetTopic(),
							MsgCapsule: &msgCapsule,
						}
					}
				}()
			}
		}
	}()

	// res stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// receive res msgs from channel and send
			err := stream.SendMsg(<-resCh)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	return nil
}
