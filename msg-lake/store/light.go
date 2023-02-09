package store

import (
	"context"
	"fmt"
	"sync"
)

const (
	ProducerChanBuffSize = 10000
	ConsumerChanBuffSize = 100
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

		producerChan:  make(MsgCapsuleChan, ProducerChanBuffSize),
		consumerChans: make(map[string]MsgCapsuleChan),
	}
}

func (box *MsgBoxLight) GetProducerChan() MsgCapsuleChan {
	return box.producerChan
}

func (box *MsgBoxLight) SetConsumerChan(consumerID string) (MsgCapsuleChan, error) {
	consumerChan := make(MsgCapsuleChan, ConsumerChanBuffSize)
	errorChan := make(chan error)
	defer close(errorChan)
	box.setConsumerChan <- setConsumerChan{
		consumerID:   consumerID,
		consumerChan: consumerChan,
		errorChan:    errorChan,
	}
	err := <-errorChan
	if err != nil {
		return nil, err
	}
	return consumerChan, nil
}

func (box *MsgBoxLight) CloseConsumerChan(consumerID string) {
	box.closeConsumerChan <- closeConsumerChan{consumerID: consumerID}
}

func (box *MsgBoxLight) Relay(ctx context.Context) {
	var (
		setConsumerChan   setConsumerChan
		closeConsumerChan closeConsumerChan
		errorChan         chan error
	)
	if box.isRelaying {
		return
	}
	box.isRelaying = true
	for {
		select {
		// handling done ctx
		case <-ctx.Done():
			for consumerID, consumerChan := range box.consumerChans {
				close(consumerChan)
				delete(box.consumerChans, consumerID)
			}
			box.isRelaying = false
			return

		// handling operation: setConsumerChan
		case setConsumerChan = <-box.setConsumerChan:
			consumerID := setConsumerChan.consumerID
			consumerChan := setConsumerChan.consumerChan
			errorChan = setConsumerChan.errorChan
			if _, exist := box.consumerChans[consumerID]; exist {
				errorChan <- fmt.Errorf("found existing consumer chan for consumer id(%s)", consumerID)
				continue
			}
			box.consumerChans[consumerID] = consumerChan
			errorChan <- nil

		// handling operation: closeConsumerChan
		case closeConsumerChan = <-box.closeConsumerChan:
			consumerID := closeConsumerChan.consumerID
			consumerChan, exist := box.consumerChans[consumerID]
			if !exist {
				continue
			}
			close(consumerChan)
			delete(box.consumerChans, consumerID)

		// handling msg
		case msgCapsule := <-box.producerChan:
			for _, consumerChan := range box.consumerChans {
				consumerChan <- msgCapsule
			}
		}
	}
}
