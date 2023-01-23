package store

import (
	"fmt"
	"sync"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreLight struct {
	msgBoxes *sync.Map // <msg_box_id>#<consumer_id>:<consumer_chans>
}

func NewMsgStoreLight() *MsgStoreLight {
	return &MsgStoreLight{
		msgBoxes: &sync.Map{},
	}
}

func (store *MsgStoreLight) Produce(msgBoxID string, msgCapsule *pb.MsgCapsule) error {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, &sync.Map{})
	msgBox := value.(*sync.Map)
	msgBox.Range(func(key, value any) bool {
		value.(chan *pb.MsgCapsule) <- msgCapsule
		return true
	})
	return nil
}

func (store *MsgStoreLight) Consume(msgBoxID, consumerID string) (<-chan *pb.MsgCapsule, error) {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, &sync.Map{})
	msgBox := value.(*sync.Map)
	if _, exist := msgBox.Load(consumerID); exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	consumerChan := make(chan *pb.MsgCapsule)
	msgBox.Store(consumerID, consumerChan)
	return consumerChan, nil
}

func (store *MsgStoreLight) Sync(msgBoxID, consumerID string) error {
	return nil
}

func (store *MsgStoreLight) Stop(msgBoxID, consumerID string) error {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	msgBox := value.(*sync.Map)
	msgBox.Delete(consumerID)
	return nil
}
