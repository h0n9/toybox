package store

import (
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

func (box *MsgBox) AppendMsg(msg *proto.Msg) error {
	box.msgs = append(box.msgs, msg)
	return nil
}

func (box *MsgBox) IncrementConsumerOffset(consumerID string) {
	value, _ := box.consumerOffsets.LoadOrStore(consumerID, 0)
	box.consumerOffsets.Store(consumerID, value.(int)+1)
}

func (box *MsgBox) CreateConsumerChan(consumerID string) chan *proto.Msg {
	consumerChan, exist := box.consumerChans[consumerID]
	if !exist {
		consumerChan = make(chan *proto.Msg)
		box.consumerChans[consumerID] = consumerChan
	}
	return consumerChan
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
	err := msgBox.AppendMsg(msg)
	if err != nil {
		return err
	}

	for consumerID, consumerChan := range msgBox.consumerChans {
		// 2. distribute msgs to consumers
		consumerChan <- msg
		// 3. update consumer's offset
		msgBox.IncrementConsumerOffset(consumerID)
	}

	return nil
}

func (store *MsgStoreMemory) Consume(msgBoxID, consumerID string) (chan *proto.Msg, error) {
	msgBox, exist := store.msgBoxes[msgBoxID]
	if !exist {
		msgBox = NewMsgBox()
		store.msgBoxes[msgBoxID] = msgBox
	}

	// 1. create consumer channel
	consumerChan := msgBox.CreateConsumerChan(consumerID)

	return consumerChan, nil
}
