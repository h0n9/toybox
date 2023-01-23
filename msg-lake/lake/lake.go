package lake

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
)

type LakeServer struct {
	pb.UnimplementedLakeServer

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

func (ls *LakeServer) Send(ctx context.Context, req *pb.SendReq) (*pb.SendRes, error) {
	err := ls.msgStore.Produce(req.GetMsgBoxId(), req.GetMsgCapsule())
	if err != nil {
		return nil, err
	}
	return &pb.SendRes{Ok: true}, nil
}

func (ls *LakeServer) Recv(req *pb.RecvReq, stream pb.Lake_RecvServer) error {
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
				err = stream.Send(&pb.RecvRes{MsgCapsule: msgCapsule})
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

	wg.Wait()

	return nil
}
