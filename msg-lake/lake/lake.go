package lake

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/relayer"
)

type LakeService struct {
	proto.UnimplementedLakeServer

	ctx     context.Context
	relayer *relayer.Relayer
}

func NewLakeService(ctx context.Context) (*LakeService, error) {
	relayer, err := relayer.NewRelayer(ctx, "0.0.0.0", 7733)
	if err != nil {
		return nil, err
	}
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

func (lakeService *LakeService) PubSub(stream proto.Lake_PubSubServer) error {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	// stream handler
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
			case proto.PubSubReqType_PUBSUB_REQ_PUBLISH:
				// TODO: publish msg to relayer
			case proto.PubSubReqType_PUBSUB_REQ_SUBSCRIBE:
				// TODO: subscribe to topic_id
			}
		}
	}()

	// pubsub handler
	wg.Add(1)
	go func() {
		defer wg.Done()
	}()

	return nil
}
