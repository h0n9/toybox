package center

import (
	"context"
	"sync"

	"github.com/h0n9/toybox/msg-lake/msg/box"
)

type Light struct {
	ctx      context.Context
	msgBoxes *sync.Map // <msg_box_id>:<msg_box>
}

func NewLight(ctx context.Context) *Light {
	return &Light{
		ctx:      ctx,
		msgBoxes: &sync.Map{},
	}
}

func (store *Light) GetMsgBox(msgBoxID string) *box.Light {
	value, exist := store.msgBoxes.Load(msgBoxID)
	if exist {
		return value.(*box.Light)
	}
	msgBoxLight := box.NewLight()
	store.msgBoxes.Store(msgBoxID, msgBoxLight)
	go msgBoxLight.Relay(store.ctx)
	return msgBoxLight
}
