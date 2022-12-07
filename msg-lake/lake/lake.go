package lake

import (
	"container/list"
	"context"
	"fmt"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type LakeServer struct {
	pb.UnimplementedLakeServer

	msgs map[string]*list.List
}

func NewLakeServer() *LakeServer {
	return &LakeServer{
		msgs: map[string]*list.List{},
	}
}

func (ls *LakeServer) Close() {
	// TODO: implement Close() method
}

func (ls *LakeServer) Send(ctx context.Context, req *pb.SendReq) (*pb.SendRes, error) {
	id := req.GetId()
	msg := req.GetMsg()

	if _, exist := ls.msgs[id]; !exist {
		ls.msgs[id] = list.New()
	}

	ls.msgs[id].PushBack(msg)

	return &pb.SendRes{Ok: true}, nil
}

func (ls *LakeServer) Recv(req *pb.RecvReq, stream pb.Lake_RecvServer) error {
	id := req.GetId()

	msgs, exist := ls.msgs[id]
	if !exist {
		return fmt.Errorf("failed to find msgs corresponding to id(%s)", req.GetId())
	}

	for msgs.Len() > 0 {
		front := msgs.Front()
		msg := front.Value.(*pb.Msg)
		err := stream.Send(&pb.RecvRes{Msg: msg})
		if err != nil {
			fmt.Println(err)
			continue
		}
		msgs.Remove(front)
	}

	return nil
}
