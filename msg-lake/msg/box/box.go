package box

import "github.com/h0n9/toybox/msg-lake/msg"

const (
	SetConsumerChanBuffSize    = 10
	DeleteConsumerChanBuffSize = 10
	ProducerChanBuffSize       = 10000
	ConsumerChanBuffSize       = 100
)

type setConsumerChan struct {
	consumerID   string
	consumerChan msg.CapsuleChan
	errorChan    chan error
}

type deleteConsumerChan struct {
	consumerID string
	errorChan  chan error
}
