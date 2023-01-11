package lake

import (
	"context"
	"fmt"
	"sync"

	"github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LakeServer struct {
	proto.UnimplementedLakeServer

	msgStore store.MsgStore
}

func NewLakeServer(msgStore store.MsgStore) *LakeServer {
	return &LakeServer{
		msgStore: msgStore,
	}
}

func (ls *LakeServer) Close() {
	// TODO: implement Close() method
}

func (ls *LakeServer) Send(ctx context.Context, req *proto.SendReq) (*proto.SendRes, error) {
	err := ls.msgStore.Produce(req.GetMsgBoxId(), req.GetMsgCapsule())
	if err != nil {
		return nil, err
	}
	return &proto.SendRes{Ok: true}, nil
}

func (ls *LakeServer) Recv(req *proto.RecvReq, stream proto.Lake_RecvServer) error {
	msgBoxID := req.GetMsgBoxId()
	consumerID := req.GetConsumerId()

	consumerChan, err := ls.msgStore.Consume(msgBoxID, consumerID)
	if err != nil {
		return err
	}

	// init wait group
	wg := sync.WaitGroup{}

	// stream msgCapsules
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stream.Context().Done():
				err := ls.msgStore.Stop(msgBoxID, consumerID)
				if err != nil {
					fmt.Println(err)
				}
				return
			case msgCapsule := <-consumerChan:
				err = stream.Send(&proto.RecvRes{MsgCapsule: msgCapsule})
				if err != nil {
					code := status.Code(err)
					if code != codes.Canceled && code != codes.Unavailable {
						fmt.Println(err)
					}
					continue
				}
			}
		}
	}()

	// sync msgCapsules
	err = ls.msgStore.Sync(msgBoxID, consumerID)
	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}
