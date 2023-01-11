package store

import (
	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStore interface {
	Produce(msgBoxID string, msgCapsule *proto.MsgCapsule) error
	Consume(msgBoxID, consumerID string) (<-chan *proto.MsgCapsule, error)
	Sync(msgBoxID, consumerID string) error
	Stop(msgBoxID, consumerID string) error
}
