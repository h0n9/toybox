package store

import (
	"sync"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStore interface {
	Produce(msgBoxID string) (*sync.Map, error)
	Consume(msgBoxID, consumerID string) (<-chan *pb.MsgCapsule, error)
	Sync(msgBoxID, consumerID string) error
	Stop(msgBoxID, consumerID string) error
}

type MsgCapsuleChan chan *pb.MsgCapsule
