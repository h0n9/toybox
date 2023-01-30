package store

import (
	"context"
	"fmt"
	"sync"

	pb "github.com/h0n9/toybox/msg-lake/proto"
)

type MsgStoreLight struct {
	ctx      context.Context
	msgBoxes *sync.Map // <msg_box_id>:<msg_box>
}

func NewMsgStoreLight(ctx context.Context) *MsgStoreLight {
	return &MsgStoreLight{
		ctx:      ctx,
		msgBoxes: &sync.Map{},
	}
}

func (store *MsgStoreLight) GetMsgBox(msgBoxID string) *MsgBoxLight {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if exist {
		return value.(*MsgBoxLight)
	}
	msgBoxLight := NewMsgBoxLight()
	store.msgBoxes.Store(msgBoxID, msgBoxLight)
	go msgBoxLight.Relay(store.ctx)
	return msgBoxLight
}

type setConsumerChan struct {
	consumerID   string
	consumerChan MsgCapsuleChan
	errorChan    chan error
}

type closeConsumerChan struct {
	consumerID string
	errorChan  chan error
}

type MsgBoxLight struct {
	isRelaying bool

	// chans for operations
	setConsumerChan   chan setConsumerChan
	closeConsumerChan chan closeConsumerChan

	// chans for msgs
	producerChan  MsgCapsuleChan
	consumerChans map[string]MsgCapsuleChan
}

func NewMsgBoxLight() *MsgBoxLight {
	return &MsgBoxLight{
		isRelaying: false,

		setConsumerChan:   make(chan setConsumerChan, 10),
		closeConsumerChan: make(chan closeConsumerChan, 10),

		producerChan:  make(MsgCapsuleChan, 1000),
		consumerChans: make(map[string]MsgCapsuleChan),
	}
}

func (box *MsgBoxLight) GetProducerChan() MsgCapsuleChan {
	return box.producerChan
}

func (box *MsgBoxLight) SetConsumerChan(consumerID string, consumerChan MsgCapsuleChan, errorChan chan error) {
	box.setConsumerChan <- setConsumerChan{
		consumerID:   consumerID,
		consumerChan: consumerChan,
		errorChan:    errorChan,
	}
}

func (box *MsgBoxLight) CloseConsumerChan(consumerID string, errorChan chan error) {
	box.closeConsumerChan <- closeConsumerChan{
		consumerID: consumerID,
		errorChan:  errorChan,
	}
}

func (box *MsgBoxLight) Relay(ctx context.Context) {
	fmt.Println("relaying")
	var (
		setConsumerChan   setConsumerChan
		closeConsumerChan closeConsumerChan
		errorChan         chan error

		consumerID   string
		consumerChan MsgCapsuleChan
		msgCapsule   *pb.MsgCapsule
	)
	if box.isRelaying {
		return
	}
	box.isRelaying = true
	for {
		select {
		// handling done ctx
		case <-ctx.Done():
			fmt.Println("done")
			for consumerID, consumerChan = range box.consumerChans {
				close(consumerChan)
				delete(box.consumerChans, consumerID)
			}
			box.isRelaying = false
			return

		// handling operation: setConsumerChan
		case setConsumerChan = <-box.setConsumerChan:
			fmt.Println("setConsumerChan")
			consumerID = setConsumerChan.consumerID
			consumerChan = setConsumerChan.consumerChan
			errorChan = setConsumerChan.errorChan
			if _, exist := box.consumerChans[consumerID]; exist {
				errorChan <- fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
				continue
			}
			box.consumerChans[consumerID] = consumerChan
			errorChan <- nil

		// handling operation: closeConsumerChan
		case closeConsumerChan = <-box.closeConsumerChan:
			fmt.Println("closeConsumerChan")
			consumerID = closeConsumerChan.consumerID
			// errorChan = closeConsumerChan.errorChan
			consumerChan, exist := box.consumerChans[consumerID]
			if !exist {
				// errorChan <- fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
				continue
			}
			close(consumerChan)
			delete(box.consumerChans, consumerID)
			// errorChan <- nil

		// handling msg
		case msgCapsule = <-box.producerChan:
			for _, consumerChan := range box.consumerChans {
				consumerChan <- msgCapsule
			}
		}
	}
}

// func (box *MsgBoxLight) CreateMsgCapsuleChan(consumerID string) (MsgCapsuleChan, error) {
// 	if _, exist := box.msgCapsuleChans.Load(consumerID); exist {
// 		return nil, fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
// 	}
// 	msgCapsuleChan := make(MsgCapsuleChan, 1)
// 	box.msgCapsuleChans.Store(consumerID, msgCapsuleChan)
// 	return msgCapsuleChan, nil
// }
//
// func (box *MsgBoxLight) RemoveMsgCapsuleChan(consumerID string) error {
// 	box.msgCapsuleChans.Delete(consumerID)
// 	return nil
// }
//
// func (box *MsgBoxLight) GetMsgCapsuleChan(consumerID string) (MsgCapsuleChan, error) {
// 	value, exist := box.msgCapsuleChans.Load(consumerID)
// 	if !exist {
// 		return nil, fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
// 	}
// 	return value.(MsgCapsuleChan), nil
// }
//
// func (box *MsgBoxLight) SendMsgCapsule(msgCapsule *pb.MsgCapsule) {
// 	box.msgCapsuleChans.Range(func(key, value any) bool {
// 		value.(MsgCapsuleChan) <- msgCapsule
// 		return true
// 	})
// }
//
