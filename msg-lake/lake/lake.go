package lake

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/relayer"
	"github.com/h0n9/toybox/msg-lake/util"
)

const (
	MaxTopicIDLen = 30
	MinTopicIDLen = 1

	RandomSubscriberIDLen = 10
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
func (lakeService *LakeService) Subscribe(req *pb.SubscribeReq, stream pb.Lake_SubscribeServer) error {
	// get parameters
	topicID := req.GetTopicId()

	// set subscribe res
	res := pb.SubscribeRes{
		Type:    pb.SubscribeResType_SUBSCRIBE_RES_TYPE_ACK,
		TopicId: topicID,
		Res: &pb.SubscribeRes_Ok{
			Ok: false,
		},
	}

	// check constraints
	if !util.CheckStrLen(topicID, MinTopicIDLen, MaxTopicIDLen) {
		err := stream.Send(&res)
		if err != nil {
			return err
		}
		return nil
	}

	// get msg center
	msgCenter := lakeService.relayer.GetMsgCenter()

	// get msg box
	msgBox, err := msgCenter.GetBox(topicID)
	if err != nil {
		err := stream.Send(&res)
		if err != nil {
			return err
		}
		return nil
	}

	// generate random subscriber id
	subscriberID := util.GenerateRandomBase64String(RandomSubscriberIDLen)

	// register subscriber id to msg box
	subscriberCh, err := msgBox.Subscribe(subscriberID)
	if err != nil {
		err := stream.Send(&res)
		if err != nil {
			return err
		}
		return nil
	}

	// update subscriber res
	res.SubscriberId = subscriberID
	res.Res = &pb.SubscribeRes_Ok{Ok: true}

	// send subscriber res
	err = stream.Send(&res)
	if err != nil {
		return err
	}

	// update subscriber res
	res.Type = pb.SubscribeResType_SUBSCRIBE_RES_TYPE_RELAY

	// relay msgs to susbscriber
	for data := range subscriberCh {
		res.Res = &pb.SubscribeRes_Data{Data: data}
		err := stream.Send(&res)
		if err != nil {
			err := msgBox.StopSubscription(subscriberID)
			if err != nil {
				lakeService.logger.Err(err).Msg("")
			}
		}
		continue
	}

	return nil
}
