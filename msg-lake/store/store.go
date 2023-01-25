package store

import (
	"sync"

	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStore interface {
	Produce(msgBoxID string) (*sync.Map, error)
	Consume(msgBoxID, consumerID string) (<-chan *proto.MsgCapsule, error)
	Sync(msgBoxID, consumerID string) error
	Stop(msgBoxID, consumerID string) error
}
