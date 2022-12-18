package store

import (
	"fmt"

	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgBox struct {
	msgs      []*proto.Msg
	consumers map[string]int // <consumer_id>:<consumer_offset>
}

func NewMsgBox() *MsgBox {
	return &MsgBox{
		msgs:      make([]*proto.Msg, 0),
		consumers: make(map[string]int),
	}
}

func (box *MsgBox) Len() int {
	return len(box.msgs)
}

func (box *MsgBox) Append(msg *proto.Msg) error {
	box.msgs = append(box.msgs, msg)
	return nil
}

func (box *MsgBox) Get(consumerID string) (*proto.Msg, error) {
	// get consumer offset
	consumerOffset, exist := box.consumers[consumerID]
	if !exist {
		consumerOffset = 0
		box.consumers[consumerID] = consumerOffset
	}

	// check constraints
	if consumerOffset > len(box.msgs) {
		return nil, fmt.Errorf("offset cannot exceed length of msg box")
	}

	// update consumer offset
	if consumerOffset+1 <= len(box.msgs) {
		box.consumers[consumerID] = consumerOffset + 1
	}

	return box.msgs[consumerOffset], nil
}

type MsgStoreMemory struct {
	msgBoxes map[string]*MsgBox // <msg_box_id>:<msg_box>
}

func NewMsgStoreMemory() *MsgStoreMemory {
	return &MsgStoreMemory{
		msgBoxes: make(map[string]*MsgBox),
	}
}

func (store *MsgStoreMemory) Push(msgBoxID string, msg *proto.Msg) error {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		msgBox = NewMsgBox()
		store.msgBoxes[msgBoxID] = msgBox
	}
	return msgBox.Append(msg)
}

func (store *MsgStoreMemory) Pop(msgBoxID, consumerID string) (*proto.Msg, error) {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		return nil, fmt.Errorf("failed to find msg box corresponding to id(%s)", msgBoxID)
	}
	return msgBox.Get(consumerID)
}

func (store *MsgStoreMemory) Len(msgBoxID string) int {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		return 0
	}
	return msgBox.Len()
}
