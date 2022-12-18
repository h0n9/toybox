package store

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/h0n9/toybox/msg-lake/proto"
)

func TestMsgStoreMemory(t *testing.T) {
	// const
	const (
		numOfMsgs        int    = 1000
		numOfConsumers   int    = 10000
		msgBoxID         string = "test"
		consumerIDPrefix string = "test-consumer"
	)

	ms := NewMsgStoreMemory()
	assert.Equal(t, 0, len(ms.msgBoxes))
	assert.Equal(t, 0, ms.Len(msgBoxID))

	for i := 0; i < numOfMsgs; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(rand.Int())))
		ms.Push(msgBoxID, &proto.Msg{
			From: &proto.Address{Address: fmt.Sprintf("addr-%d", i)},
			Data: &proto.Data{Data: hash[:]},
		})
	}

	assert.Equal(t, 1, len(ms.msgBoxes))
	assert.Equal(t, numOfMsgs, ms.Len(msgBoxID))
	assert.Equal(t, 0, len(ms.msgBoxes[msgBoxID].consumers))

	for i := 0; i < numOfConsumers; i++ {
		for j := 0; j < numOfMsgs; j++ {
			msg, err := ms.Pop(msgBoxID, fmt.Sprintf("%s-%d", consumerIDPrefix, i))
			assert.NoError(t, err)
			assert.Equal(t, fmt.Sprintf("addr-%d", j), msg.GetFrom().GetAddress())
		}
	}

	assert.Equal(t, 1, len(ms.msgBoxes))
	assert.Equal(t, numOfMsgs, ms.Len(msgBoxID))
	assert.Equal(t, numOfConsumers, len(ms.msgBoxes[msgBoxID].consumers))
}
