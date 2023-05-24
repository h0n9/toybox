package lake

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/rs/zerolog"

	"github.com/h0n9/toybox/msg-lake/msg"
	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/relayer"
	"github.com/h0n9/toybox/msg-lake/util"
)

const (
	MaxTopicIDLen = 30
	MinTopicIDLen = 1
)

type LakeService struct {
	pb.UnimplementedLakeServer

	ctx     context.Context
	logger  *zerolog.Logger
	relayer *relayer.Relayer
}

func NewLakeService(ctx context.Context, logger *zerolog.Logger) (*LakeService, error) {
	subLogger := logger.With().Str("module", "lake-service").Logger()
	relayer, err := relayer.NewRelayer(ctx, logger, "0.0.0.0", 7733)
	if err != nil {
		return nil, err
	}
	go relayer.DiscoverPeers()
	return &LakeService{
		ctx:     ctx,
		logger:  &subLogger,
		relayer: relayer,
	}, nil
}

func (lakeService *LakeService) Close() {
	if lakeService.relayer != nil {
		lakeService.relayer.Close()
	}
	lakeService.logger.Info().Msg("closed lake service")
}

func (lakeService *LakeService) Publish(ctx context.Context, req *pb.PublishReq) (*pb.PublishRes, error) {
	// get parameters
	topicID := req.GetTopicId()
	data := req.GetData()

	// set publish res
	publishRes := pb.PublishRes{
		TopicId: topicID,
		Ok:      false,
	}

	// check constraints
	if !util.CheckStrLen(topicID, MinTopicIDLen, MaxTopicIDLen) {
		return &publishRes, fmt.Errorf("failed to verify length of topic id")
	}

	// get msg center
	msgCenter := lakeService.relayer.GetMsgCenter()

	// get msg box
	msgBox, err := msgCenter.GetBox(topicID)
	if err != nil {
		return &publishRes, err
	}

	// publish msg
	err = msgBox.Publish(data)
	if err != nil {
		return &publishRes, err
	}

	// update publish res
	publishRes.Ok = true

	return &publishRes, nil
}
func (lakeService *LakeService) Subscribe(stream pb.Lake_SubscribeServer) error

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
						lakeService.logger.Err(err).Msg("")
					}
				}
				lakeService.logger.Info().Str("subscriber", subscriberID).Msg("left")
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
				lakeService.logger.Info().Str("subscriber", subscriberID).Msg("subscribing")
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
					lakeService.logger.Err(err).Msg("")
					continue
				}
			case data := <-subscriberCh:
				err := stream.Send(&pb.PubSubRes{
					Type: pb.PubSubResType_PUB_SUB_RES_TYPE_SUBSCRIBE,
					Data: data,
				})
				if err != nil {
					lakeService.logger.Err(err).Msg("")
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
