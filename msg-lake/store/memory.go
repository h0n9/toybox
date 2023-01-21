package store

import (
	"fmt"
	"sync"
	"sync/atomic"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreMemory struct {
	msgBoxes *sync.Map // <msg_box_id>:<msg_box>
}

func NewMsgStoreMemory() *MsgStoreMemory {
	return &MsgStoreMemory{
		msgBoxes: &sync.Map{},
	}
}

func (store *MsgStoreMemory) Produce(msgBoxID string, msgCapsule *pb.MsgCapsule) error {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, NewMsgBox())
	msgBox := value.(*MsgBox)

	// 1. store msg
	offset := msgBox.Append(msgCapsule)

	msgBox.consumerChans.Range(func(key, value any) bool {
		// 2. distribute msgs to consumers
		value.(chan *pb.MsgCapsule) <- msgCapsule
		// 3. update consumer offset
		msgBox.consumerOffsets.Store(key.(string), offset)
		return true
	})

	return nil
}

func (store *MsgStoreMemory) Consume(msgBoxID, consumerID string) (<-chan *pb.MsgCapsule, error) {
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
	for ; consumerOffset < backOffset; consumerOffset++ {
		value, exist := msgBox.msgCapsules.Load(consumerOffset)
		if !exist {
			continue
		}
		consumerChan <- value.(*pb.MsgCapsule)
		msgBox.consumerOffsets.Store(consumerID, consumerOffset)
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
