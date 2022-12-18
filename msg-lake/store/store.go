package store

import (
	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStore interface {
	Push(id string, msg *pb.Msg) error
	Pop(id, consumer string) (*pb.Msg, error)
	Len(id string) int
	Behind(id, consumer string) int
}
