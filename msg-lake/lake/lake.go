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

func (ls *LakeServer) Send(stream pb.Lake_SendServer) error {
	var (
		msgBoxID string = ""
		msgBox   *sync.Map
	)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SendRes{Ok: true})
		}
		if err != nil {
			return err
		}
		if msgBoxID != req.GetMsgBoxId() {
			newMsgBox, err := ls.msgStore.Produce(req.GetMsgBoxId())
			if err != nil {
				return err
			}
			msgBoxID = req.GetMsgBoxId()
			msgBox = newMsgBox
		}
		msgBox.Range(func(key, value any) bool {
			value.(chan *pb.MsgCapsule) <- req.GetMsgCapsule()
			return true
		})
	}
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
