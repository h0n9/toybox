package lake

import (
	"context"
	"fmt"

	"github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
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
	err := ls.msgStore.Push(req.GetMsgBoxId(), req.GetMsg())
	if err != nil {
		return nil, err
	}
	return &proto.SendRes{Ok: true}, nil
}

func (ls *LakeServer) Recv(req *proto.RecvReq, stream proto.Lake_RecvServer) error {
	msgBoxID := req.GetMsgBoxId()
	consumerID := req.GetConsumerId()
	for ls.msgStore.Behind(msgBoxID, consumerID) > 0 {
		msg, err := ls.msgStore.Pop(msgBoxID, consumerID)
		if err != nil {
			return err
		}
		err = stream.Send(&proto.RecvRes{Msg: msg})
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	return nil
}
