package store

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/postie-labs/go-postie-lib/crypto"
	"google.golang.org/protobuf/proto"

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
	// 0. verify msg signature
	msg := msgCapsule.GetMsg()
	signature := msgCapsule.GetSignature()
	pubKey, err := crypto.GenPubKeyFromBytes(signature.GetPubKey())
	if err != nil {
		return err
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	sigBytes := signature.GetSigBytes()
	if !pubKey.Verify(msgBytes, sigBytes) {
		return fmt.Errorf("failed to verify msg signature")
	}

	value, _ := store.msgBoxes.LoadOrStore(msgBoxID, NewMsgBox())
	msgBox := value.(*MsgBoxMemory)

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
	msgBox := value.(*MsgBoxMemory)
	return msgBox.CreateConsumerChan(consumerID)
}

func (store *MsgStoreMemory) Sync(msgBoxID, consumerID string) error {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if !exist {
		return fmt.Errorf("failed to find msg box for id(%s)", msgBoxID)
	}
	msgBox := value.(*MsgBoxMemory)
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
	msgBox := value.(*MsgBoxMemory)
	return msgBox.RemoveConsumerChan(consumerID)
}

type MsgBoxMemory struct {
	frontOffset     uint64
	backOffset      uint64
	msgCapsules     *sync.Map // <offset>:<msg_capsule>
	consumerChans   *sync.Map // <consumer_id>:<consumer_chan>
	consumerOffsets *sync.Map // <consumer_id>:<consumer_offset>
}

func NewMsgBox() *MsgBoxMemory {
	return &MsgBoxMemory{
		frontOffset:     0,
		backOffset:      0,
		msgCapsules:     &sync.Map{},
		consumerChans:   &sync.Map{},
		consumerOffsets: &sync.Map{},
	}
}

func (box *MsgBoxMemory) Append(msgCapsule *pb.MsgCapsule) uint64 {
	offset := atomic.LoadUint64(&box.backOffset)
	box.msgCapsules.Store(offset, msgCapsule)
	offset = atomic.AddUint64(&box.backOffset, 1)
	return offset
}

func (box *MsgBoxMemory) Len() uint64 {
	frontOffset := atomic.LoadUint64(&box.frontOffset)
	backOffset := atomic.LoadUint64(&box.backOffset)
	if frontOffset == backOffset {
		return 0
	}
	return backOffset - frontOffset + 1
}

func (box *MsgBoxMemory) CreateConsumerChan(consumerID string) (chan *pb.MsgCapsule, error) {
	_, exist := box.consumerChans.Load(consumerID)
	if exist {
		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
	}
	consumerChan := make(chan *pb.MsgCapsule)
	box.consumerChans.Store(consumerID, consumerChan)
	return consumerChan, nil
}

func (box *MsgBoxMemory) RemoveConsumerChan(consumerID string) error {
	box.consumerChans.Delete(consumerID)
	return nil
}

func (box *MsgBoxMemory) GetConsumerChan(consumerID string) (chan *pb.MsgCapsule, error) {
	value, exist := box.consumerChans.Load(consumerID)
	if !exist {
		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
	}
	return value.(chan *pb.MsgCapsule), nil
}
