package lake

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"google.golang.org/protobuf/proto"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/relayer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
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

	resCh := make(chan pb.PubSubRes)

	port := rand.Intn(8080-1000+1) + 1000
	h, err := lakeService.relayer.NewSubHost(port)
	if err != nil {
		return err
	}
	ps, err := pubsub.NewGossipSub(stream.Context(), h)
	if err != nil {
		return err
	}
	tm := NewTopicManager(stream.Context(), ps, resCh)
	defer func() {
		tm.Close()
		h.Close()
		wg.Wait()
	}()

	// req stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			pubSubReq, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				fmt.Println(err)
				return
			}

			switch pubSubReq.Type {
			case pb.PubSubReqType_PUB_SUB_REQ_TYPE_PUBLISH:
				topic, err := tm.Join(pubSubReq.TopicId)
				if err != nil {
					resCh <- pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				data, err := proto.Marshal(pubSubReq.GetMsgCapsule())
				if err != nil {
					resCh <- pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				err = topic.Publish(stream.Context(), data)
				if err != nil {
					resCh <- pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				resCh <- pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
					Ok:   true,
				}
			case pb.PubSubReqType_PUB_SUB_REQ_TYPE_SUBSCRIBE:
				_, err = tm.Join(pubSubReq.TopicId)
				if err != nil {
					resCh <- pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
						Ok:   false,
					}
					continue
				}
				resCh <- pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
					Ok:   true,
				}
			}
		}
	}()

	// res stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			// receive res msgs from channel and send
			res := <-resCh
			fmt.Println(res)
			err := stream.Send(&res)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}
	}()

	return nil
}
