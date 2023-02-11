package lake

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/h0n9/toybox/msg-lake/msg"
	"github.com/h0n9/toybox/msg-lake/msg/box"
	"github.com/h0n9/toybox/msg-lake/msg/center"
	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type LakeServer struct {
	pb.UnimplementedLakeServer

	ctx       context.Context
	msgCenter *center.Light
}

func NewLakeServer(ctx context.Context) *LakeServer {
	return &LakeServer{
		ctx:       ctx,
		msgCenter: center.NewLight(ctx),
	}
}

func (ls *LakeServer) Close() {
	// TODO: implement Close() method
}

func (ls *LakeServer) Send(stream pb.Lake_SendServer) error {
	var (
		msgBoxID string = ""
		msgBox   *box.Light

		producerChan msg.CapsuleChan
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
			msgBox = ls.msgCenter.GetMsgBox(msgBoxID)
			producerChan = msgBox.GetProducerChan()
		}
		producerChan <- req.GetMsgCapsule()
	}
}

func (ls *LakeServer) Recv(req *pb.RecvReq, stream pb.Lake_RecvServer) error {
	msgBoxID := req.GetMsgBoxId()
	consumerID := req.GetConsumerId()

	msgBox := ls.msgCenter.GetMsgBox(msgBoxID)
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
