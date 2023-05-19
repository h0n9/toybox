package lake

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/h0n9/toybox/msg-lake/msg"
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

	var (
		resCh chan *pb.PubSubRes = make(chan *pb.PubSubRes)

		subscriberID string
		subscriberCh msg.SubscriberCh

		msgBoxes map[string]*msg.Box = make(map[string]*msg.Box)
	)
	defer close(resCh)

	msgCenter := lakeService.relayer.GetMsgCenter()

	// req stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			pubSubReq, err := stream.Recv()
			if err != nil {
				for _, msgBox := range msgBoxes {
					err = msgBox.StopSubscription(subscriberID)
					if err != nil {
						fmt.Println(err)
					}
				}
				fmt.Printf("subscriber '%s' left\n", subscriberID)
				return
			}

			switch pubSubReq.Type {
			case pb.PubSubReqType_PUB_SUB_REQ_TYPE_PUBLISH:
				msgBox, exist := msgBoxes[pubSubReq.TopicId]
				if !exist {
					newMsgBox, err := msgCenter.GetBox(pubSubReq.TopicId)
					if err != nil {
						resCh <- &pb.PubSubRes{
							Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
							Ok:   false,
						}
						continue
					}
					msgBox = newMsgBox
					msgBoxes[pubSubReq.TopicId] = msgBox
				}
				err = msgBox.Publish(pubSubReq.GetData())
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
				msgBox, exist := msgBoxes[pubSubReq.TopicId]
				if !exist {
					newMsgBox, err := msgCenter.GetBox(pubSubReq.TopicId)
					if err != nil {
						resCh <- &pb.PubSubRes{
							Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
							Ok:   false,
						}
						continue
					}
					msgBox = newMsgBox
					msgBoxes[pubSubReq.TopicId] = msgBox
				}
				tmpSubscriberID := fmt.Sprintf("%s-%s",
					pubSubReq.GetSubscriberId(),
					generateRandomBase64String(4),
				)
				subscriberCh, err = msgBox.Subscribe(tmpSubscriberID)
				if err != nil {
					resCh <- &pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
						Ok:   false,
					}
					continue
				}
				subscriberID = tmpSubscriberID
				resCh <- &pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
					Ok:   true,
				}
				fmt.Printf("subscriber '%s' subscribing\n", subscriberID)
			}
		}
	}()

	// res stream handler
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			// receive res msgs from channel and send
			case res := <-resCh:
				err := stream.Send(res)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case data := <-subscriberCh:
				err := stream.Send(&pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
					Data: data,
				})
				if err != nil {
					fmt.Println(err)
					continue
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

func generateRandomBase64String(size int) string {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return base64.RawStdEncoding.EncodeToString(bytes)
}
