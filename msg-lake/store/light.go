package store

import (
	"fmt"
	"sync"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreLight struct {
	msgBoxes *sync.Map // <msg_box_id>:<msg_box>
}

func NewMsgStoreLight() *MsgStoreLight {
	return &MsgStoreLight{
		msgBoxes: &sync.Map{},
	}
}

func (store *MsgStoreLight) GetMsgBox(msgBoxID string) *MsgBoxLight {
	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, NewMsgBoxLight())
	return value.(*MsgBoxLight)
}

type MsgBoxLight struct {
	msgCapsuleChans *sync.Map // <consumer_id>:<msg_capsule_chan>
}

func NewMsgBoxLight() *MsgBoxLight {
	return &MsgBoxLight{msgCapsuleChans: &sync.Map{}}
}

func (box *MsgBoxLight) CreateMsgCapsuleChan(consumerID string) (MsgCapsuleChan, error) {
	if _, exist := box.msgCapsuleChans.Load(consumerID); exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	msgCapsuleChan := make(MsgCapsuleChan)
	box.msgCapsuleChans.Store(consumerID, msgCapsuleChan)
	return msgCapsuleChan, nil
}

func (box *MsgBoxLight) RemoveMsgCapsuleChan(consumerID string) error {
	box.msgCapsuleChans.Delete(consumerID)
	return nil
}

func (box *MsgBoxLight) GetMsgCapsuleChan(consumerID string) (MsgCapsuleChan, error) {
	value, exist := box.msgCapsuleChans.Load(consumerID)
	if !exist {
		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
	}
	return value.(MsgCapsuleChan), nil
}

func (box *MsgBoxLight) SendMsgCapsule(msgCapsule *pb.MsgCapsule) {
	box.msgCapsuleChans.Range(func(key, value any) bool {
		value.(MsgCapsuleChan) <- msgCapsule
		return true
	})
}
