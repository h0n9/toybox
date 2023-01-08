package store

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgBox struct {
	frontOffset     uint64
	backOffset      uint64
	msgs            *sync.Map // <offset>:<msg>
	consumerChans   *sync.Map // <consumer_id>:<consumer_chan>
	consumerOffsets *sync.Map // <consumer_id>:<consumer_offset>
}

func NewMsgBox() *MsgBox {
	return &MsgBox{
		frontOffset:     0,
		backOffset:      0,
		msgs:            &sync.Map{},
		consumerChans:   &sync.Map{},
		consumerOffsets: &sync.Map{},
	}
}

func (box *MsgBox) Append(msg *proto.Msg) uint64 {
	offset := atomic.LoadUint64(&box.backOffset)
	box.msgs.Store(offset, msg)
	atomic.AddUint64(&box.backOffset, 1)
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

func (box *MsgBox) CreateConsumerChan(consumerID string) (chan *proto.Msg, error) {
	_, exist := box.consumerChans.Load(consumerID)
	if exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	consumerChan := make(chan *proto.Msg)
	box.consumerChans.Store(consumerID, consumerChan)
	return consumerChan, nil
}

func (box *MsgBox) RemoveConsumerChan(consumerID string) error {
	value, exist := box.consumerChans.LoadAndDelete(consumerID)
	if !exist {
		return nil
	}
	close(value.(chan *proto.Msg))
	return nil
}

func (box *MsgBox) GetConsumerChan(consumerID string) (chan *proto.Msg, error) {
	value, exist := box.consumerChans.Load(consumerID)
	if !exist {
		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
	}
	return value.(chan *proto.Msg), nil
}

type MsgStoreMemory struct {
	msgBoxes *sync.Map // <msg_box_id>:<msg_box>
}

func NewMsgStoreMemory() *MsgStoreMemory {
	return &MsgStoreMemory{
		msgBoxes: &sync.Map{},
	}
}

func (store *MsgStoreMemory) Produce(msgBoxID string, msg *proto.Msg) error {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, NewMsgBox())
	msgBox := value.(*MsgBox)

	// 1. store msg
	offset := msgBox.Append(msg)

	msgBox.consumerChans.Range(func(key, value any) bool {
		select {
		// 2. distribute msgs to consumers
		case value.(chan *proto.Msg) <- msg:
			// 3. update consumer offset
			msgBox.consumerOffsets.Store(key.(string), offset)
		default:
		}
		return true
	})

	return nil
}

func (store *MsgStoreMemory) Consume(msgBoxID, consumerID string) (<-chan *proto.Msg, error) {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, NewMsgBox())
	msgBox := value.(*MsgBox)
	return msgBox.CreateConsumerChan(consumerID)
}

func (store *MsgStoreMemory) Sync(msgBoxID, consumerID string) error {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	msgBox := value.(*MsgBox)
	consumerChan, err := msgBox.GetConsumerChan(consumerID)
	if err != nil {
		return err
	}
	frontOffset := atomic.LoadUint64(&msgBox.frontOffset)
	backOffset := atomic.LoadUint64(&msgBox.backOffset)
	value, _ = msgBox.consumerOffsets.LoadOrStore(consumerID, frontOffset)
	consumerOffset := value.(uint64)
	if consumerOffset != 0 {
		consumerOffset += 1
	}
	for ; consumerOffset <= backOffset; consumerOffset++ {
		value, exist := msgBox.msgs.Load(consumerOffset)
		if !exist {
			continue
		}
		select {
		case consumerChan <- value.(*proto.Msg):
			msgBox.consumerOffsets.Store(consumerID, consumerOffset)
		default:
		}
	}
	return nil
}

func (store *MsgStoreMemory) Stop(msgBoxID, consumerID string) error {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	msgBox := value.(*MsgBox)
	return msgBox.RemoveConsumerChan(consumerID)
}
