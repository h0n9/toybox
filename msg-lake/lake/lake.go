package lake

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"

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

		msgSubscriberID string
		msgSubscribeCh  msg.SubscribeCh

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
				fmt.Println("leave all msg boxes")
				for _, msgBox := range msgBoxes {
					err = msgBox.StopSubscription(msgSubscriberID)
					if err != nil {
						fmt.Println(err)
					}
				}
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
				data, err := proto.Marshal(pubSubReq.GetMsgCapsule())
				if err != nil {
					resCh <- &pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_PUBLISH,
						Ok:   false,
					}
					continue
				}
				err = msgBox.Publish(data)
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
				msgSubscribeCh, err = msgBox.Subscribe(pubSubReq.GetSubscriberId())
				if err != nil {
					resCh <- &pb.PubSubRes{
						Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
						Ok:   false,
					}
					continue
				}
				msgSubscriberID = pubSubReq.GetSubscriberId()
				resCh <- &pb.PubSubRes{
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
			select {
			// receive res msgs from channel and send
			case res := <-resCh:
				err := stream.Send(res)
				if err != nil {
					fmt.Println(err)
					continue
				}
			case data := <-msgSubscribeCh:
				msgCapsule := pb.MsgCapsule{}
				err := proto.Unmarshal(data, &msgCapsule)
				if err != nil {
					fmt.Println(err)
					continue
				}
				err = stream.Send(&pb.PubSubRes{
					Type:       pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
					MsgCapsule: &msgCapsule,
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
