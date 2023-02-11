package box

import (
	"context"
	"fmt"

	"github.com/h0n9/toybox/msg-lake/msg"
)

type Light struct {
	isRelaying bool

	// chans for operations
	setConsumerChan    chan setConsumerChan
	deleteConsumerChan chan deleteConsumerChan

	// chans for msgs
	producerChan  msg.CapsuleChan
	consumerChans map[string]msg.CapsuleChan
}

func NewLight() *Light {
	return &Light{
		isRelaying: false,

		setConsumerChan:    make(chan setConsumerChan, SetConsumerChanBuffSize),
		deleteConsumerChan: make(chan deleteConsumerChan, DeleteConsumerChanBuffSize),

		producerChan:  make(msg.CapsuleChan, ProducerChanBuffSize),
		consumerChans: make(map[string]msg.CapsuleChan),
	}
}

func (box *Light) GetProducerChan() msg.CapsuleChan {
	return box.producerChan
}

func (box *Light) CreateConsumerChan(consumerID string) (msg.CapsuleChan, error) {
	consumerChan := make(msg.CapsuleChan, ConsumerChanBuffSize)
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

func (box *Light) CloseConsumerChan(consumerID string) error {
	errorChan := make(chan error)
	defer close(errorChan)
	box.deleteConsumerChan <- deleteConsumerChan{
		consumerID: consumerID,
		errorChan:  errorChan,
	}
	return <-errorChan
}

func (box *Light) Relay(ctx context.Context) {
	var (
		setConsumerChan    setConsumerChan
		deleteConsumerChan deleteConsumerChan
		errorChan          chan error
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

		// handling operation: deleteConsumerChan
		case deleteConsumerChan = <-box.deleteConsumerChan:
			consumerID := deleteConsumerChan.consumerID
			errorChan := deleteConsumerChan.errorChan
			consumerChan, exist := box.consumerChans[consumerID]
			if !exist {
				errorChan <- fmt.Errorf("failed to find consumer chan for consumer id(%s)", consumerID)
				continue
			}
			close(consumerChan)
			delete(box.consumerChans, consumerID)
			errorChan <- nil

		// handling msg
		case msgCapsule := <-box.producerChan:
			for _, consumerChan := range box.consumerChans {
				consumerChan <- msgCapsule
			}
		}
	}
}
