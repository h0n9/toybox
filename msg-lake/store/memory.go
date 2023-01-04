package store

import (
	"fmt"
	"sync"

	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgBox struct {
	msgs            []*proto.Msg
	consumerChans   map[string]chan *proto.Msg // <consumer_id>:<consumer_chan>
	consumerOffsets *sync.Map                  // <consumer_id>:<consumer_offset>
}

func NewMsgBox() *MsgBox {
	return &MsgBox{
		msgs:            make([]*proto.Msg, 0),
		consumerChans:   make(map[string]chan *proto.Msg, 0),
		consumerOffsets: &sync.Map{},
	}
}

func (box *MsgBox) Append(msg *proto.Msg) int {
	box.msgs = append(box.msgs, msg)
	return len(box.msgs) - 1
}

func (box *MsgBox) Len() int {
	return len(box.msgs)
}

func (box *MsgBox) CreateConsumerChan(consumerID string) (chan *proto.Msg, error) {
	_, exist := box.consumerChans[consumerID]
	if exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	consumerChan := make(chan *proto.Msg)
	box.consumerChans[consumerID] = consumerChan
	return consumerChan, nil
}

func (box *MsgBox) RemoveConsumerChan(consumerID string) error {
	consumerChan, exist := box.consumerChans[consumerID]
	if !exist {
		return nil
	}
	delete(box.consumerChans, consumerID)
	close(consumerChan)
	return nil
}

func (box *MsgBox) GetConsumerChan(consumerID string) (chan *proto.Msg, error) {
	consumerChan, exist := box.consumerChans[consumerID]
	if !exist {
		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
	}
	return consumerChan, nil
}

type MsgStoreMemory struct {
	msgBoxes map[string]*MsgBox // <msg_box_id>:<msg_box>
}

func NewMsgStoreMemory() *MsgStoreMemory {
	return &MsgStoreMemory{
		msgBoxes: make(map[string]*MsgBox),
	}
}

func (store *MsgStoreMemory) Produce(msgBoxID string, msg *proto.Msg) error {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		msgBox = NewMsgBox()
		store.msgBoxes[msgBoxID] = msgBox
	}

	// 1. store msg
	offset := msgBox.Append(msg)

	for consumerID, consumerChan := range msgBox.consumerChans {
		// 2. distribute msgs to consumers
		consumerChan <- msg
		// 3. update consumer offset
		msgBox.consumerOffsets.Store(consumerID, offset)
	}

	return nil
}

func (store *MsgStoreMemory) Consume(msgBoxID, consumerID string) (<-chan *proto.Msg, error) {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		msgBox = NewMsgBox()
		store.msgBoxes[msgBoxID] = msgBox
	}
	return msgBox.CreateConsumerChan(consumerID)
}

func (store *MsgStoreMemory) Sync(msgBoxID, consumerID string) error {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	consumerChan, err := msgBox.GetConsumerChan(consumerID)
	if err != nil {
		return err
	}
	value, _ := msgBox.consumerOffsets.LoadOrStore(consumerID, -1)
	consumerOffset := value.(int) + 1 // next offset
	len := msgBox.Len()
	for ; consumerOffset < len; consumerOffset++ {
		consumerChan <- msgBox.msgs[consumerOffset]
		msgBox.consumerOffsets.Store(consumerID, consumerOffset)
	}
	return nil
}

func (store *MsgStoreMemory) Stop(msgBoxID, consumerID string) error {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	return msgBox.RemoveConsumerChan(consumerID)
}
