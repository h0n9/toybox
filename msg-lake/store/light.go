package store

import (
	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreLight struct {
}

func NewMsgStoreLight() *MsgStoreLight {
	return &MsgStoreLight{}
}

func (store *MsgStoreLight) Produce(msgBoxID string, msgCapsule *pb.MsgCapsule) error {
	return nil
}

func (store *MsgStoreLight) Consume(msgBoxID, consumerID string) (<-chan *pb.MsgCapsule, error) {
	return nil, nil
}

func (store *MsgStoreLight) Sync(msgBoxID, consumerID string) error {
	return nil
}

func (store *MsgStoreLight) Stop(msgBoxID, consumerID string) error {
	return nil
}
