package store

import (
	"container/list"
	"fmt"

	"github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreMemory struct {
	msgs map[string]*list.List
}

func NewMsgStoreMemory() *MsgStoreMemory {
	return &MsgStoreMemory{msgs: make(map[string]*list.List)}
}

func (store *MsgStoreMemory) Push(id string, msg *proto.Msg) error {
	if _, exist := store.msgs[id]; !exist {
		store.msgs[id] = list.New()
	}
	store.msgs[id].PushBack(msg)
	return nil
}

func (store *MsgStoreMemory) Pop(id string) (*proto.Msg, error) {
	msgs, exist := store.msgs[id]
	if !exist {
		return nil, fmt.Errorf("failed to find msgs corresponding to id(%s)", id)
	}
	front := msgs.Front()
	msg := front.Value.(*proto.Msg)
	msgs.Remove(front)
	return msg, nil
}

func (store *MsgStoreMemory) Len(id string) int {
	msgs, exist := store.msgs[id]
	if !exist {
		return 0
	}
	return msgs.Len()
}
