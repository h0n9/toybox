package box

import "github.com/h0n9/toybox/msg-lake/msg"

const (
	ProducerChanBuffSize = 10000
	ConsumerChanBuffSize = 100
)

type setConsumerChan struct {
	consumerID   string
	consumerChan msg.CapsuleChan
	errorChan    chan error
}

type closeConsumerChan struct {
	consumerID string
	errorChan  chan error
}
