package lake

import (
	"context"
	"fmt"
	"io"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LakeServer struct {
	pb.UnimplementedLakeServer

	ctx      context.Context
	msgStore *store.MsgStoreLight
}

func NewLakeServer(ctx context.Context) *LakeServer {
	return &LakeServer{
		ctx:      ctx,
		msgStore: store.NewMsgStoreLight(ctx),
	}
}

func (ls *LakeServer) Close() {
	// TODO: implement Close() method
}

func (ls *LakeServer) Send(stream pb.Lake_SendServer) error {
	var (
		msgBoxID string = ""
		msgBox   *store.MsgBoxLight

		producerChan store.MsgCapsuleChan
	)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SendRes{Ok: true})
		}
		if err != nil {
			return err
		}
		newMsgBoxID := req.GetMsgBoxId()
		if msgBoxID != newMsgBoxID {
			msgBoxID = newMsgBoxID
			msgBox = ls.msgStore.GetMsgBox(msgBoxID)
			producerChan = msgBox.GetProducerChan()
		}
		producerChan <- req.GetMsgCapsule()
	}
}

func (ls *LakeServer) Recv(req *pb.RecvReq, stream pb.Lake_RecvServer) error {
	msgBoxID := req.GetMsgBoxId()
	consumerID := req.GetConsumerId()

	msgBox := ls.msgStore.GetMsgBox(msgBoxID)
	consumerChan, err := msgBox.SetConsumerChan(consumerID)
	if err != nil {
		return err
	}

	for {
		select {
		case <-stream.Context().Done():
			return msgBox.CloseConsumerChan(consumerID)
		case msgCapsule := <-consumerChan:
			err := stream.Send(&pb.RecvRes{MsgCapsule: msgCapsule})
			if err != nil {
				code := status.Code(err)
				if code != codes.Canceled && code != codes.Unavailable {
					fmt.Println(err)
				}
			}
		}
	}
}
