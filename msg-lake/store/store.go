package store

import (
	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStore interface {
	Produce(msgBoxID string, msg *proto.Msg) error
	Consume(msgBoxID, consumerID string) (<-chan *proto.Msg, error)
	Stop(msgBoxID, consumerID string) error
}
