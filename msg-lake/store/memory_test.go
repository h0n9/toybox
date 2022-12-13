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
		N  int    = 1000
		ID string = "test"
	)

	ms := NewMsgStoreMemory()
	assert.Equal(t, 0, len(ms.msgs))
	assert.Equal(t, 0, ms.Len(ID))

	for i := 0; i < N; i++ {
		hash := sha256.Sum256([]byte(strconv.Itoa(rand.Int())))
		ms.Push(ID, &proto.Msg{
			From: &proto.Address{Address: fmt.Sprintf("addr-%d", i)},
			Data: &proto.Data{Data: hash[:]},
		})
	}

	assert.Equal(t, 1, len(ms.msgs))
	assert.Equal(t, N, ms.Len(ID))

	for i := 0; i < N; i++ {
		msg, err := ms.Pop(ID)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("addr-%d", i), msg.GetFrom().GetAddress())
	}

	assert.Equal(t, 1, len(ms.msgs))
	assert.Equal(t, 0, ms.Len(ID))
}
