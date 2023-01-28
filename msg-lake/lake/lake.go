package lake

import (
	"fmt"
	"io"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h0n9/toybox/msg-lake/proto"
	"github.com/h0n9/toybox/msg-lake/store"
)

type LakeServer struct {
	pb.UnimplementedLakeServer

	msgStore *store.MsgStoreLight
}

func NewLakeServer() *LakeServer {
	return &LakeServer{
		msgStore: store.NewMsgStoreLight(),
	}
}

func (ls *LakeServer) Close() {
	// TODO: implement Close() method
}

func (ls *LakeServer) Send(stream pb.Lake_SendServer) error {
	var (
		msgBoxID string = ""
		msgBox   *store.MsgBoxLight
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
		}
		go msgBox.SendMsgCapsule(req.GetMsgCapsule())
	}
}

func (ls *LakeServer) Recv(req *pb.RecvReq, stream pb.Lake_RecvServer) error {
	msgBoxID := req.GetMsgBoxId()
	consumerID := req.GetConsumerId()

	msgBox := ls.msgStore.GetMsgBox(msgBoxID)
	msgCapsuleChan, err := msgBox.CreateMsgCapsuleChan(consumerID)
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
				msgBox.RemoveMsgCapsuleChan(consumerID)
				return
			case msgCapsule := <-msgCapsuleChan:
				err := stream.Send(&pb.RecvRes{MsgCapsule: msgCapsule})
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
