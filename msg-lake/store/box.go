package store

import (
	"fmt"
	"sync"
	"sync/atomic"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgBox struct {
	frontOffset     uint64
	backOffset      uint64
	msgCapsules     *sync.Map // <offset>:<msg_capsule>
	consumerChans   *sync.Map // <consumer_id>:<consumer_chan>
	consumerOffsets *sync.Map // <consumer_id>:<consumer_offset>
}

func NewMsgBox() *MsgBox {
	return &MsgBox{
		frontOffset:     0,
		backOffset:      0,
		msgCapsules:     &sync.Map{},
		consumerChans:   &sync.Map{},
		consumerOffsets: &sync.Map{},
	}
}

func (box *MsgBox) Append(msgCapsule *pb.MsgCapsule) uint64 {
	offset := atomic.LoadUint64(&box.backOffset)
	box.msgCapsules.Store(offset, msgCapsule)
	offset = atomic.AddUint64(&box.backOffset, 1)
	return offset
}

func (box *MsgBox) Len() uint64 {
	frontOffset := atomic.LoadUint64(&box.frontOffset)
	backOffset := atomic.LoadUint64(&box.backOffset)
	if frontOffset == backOffset {
		return 0
	}
	return backOffset - frontOffset + 1
}

func (box *MsgBox) CreateConsumerChan(consumerID string) (chan *pb.MsgCapsule, error) {
	_, exist := box.consumerChans.Load(consumerID)
	if exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	consumerChan := make(chan *pb.MsgCapsule)
	box.consumerChans.Store(consumerID, consumerChan)
	return consumerChan, nil
}

func (box *MsgBox) RemoveConsumerChan(consumerID string) error {
	box.consumerChans.Delete(consumerID)
	return nil
}

func (box *MsgBox) GetConsumerChan(consumerID string) (chan *pb.MsgCapsule, error) {
	value, exist := box.consumerChans.Load(consumerID)
	if !exist {
		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
	}
	return value.(chan *pb.MsgCapsule), nil
}
