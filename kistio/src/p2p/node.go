package p2p

import (
	"bytes"
	"context"

	"github.com/postie-labs/go-postie-lib/crypto"
)

type Node struct {
	ctx context.Context

	privKey *crypto.PrivKey
	pubKey  *crypto.PubKey
	addr    crypto.Addr
}

func NewNode(ctx context.Context, seed []byte) (*Node, error) {
	privKey, err := crypto.GenPrivKey()
	if !bytes.Equal(seed, []byte{}) {
		privKey, err = crypto.GenPrivKeyFromSeed(seed)
	}
	if err != nil {
		return nil, err
	}
	return &Node{
		ctx:     ctx,
		privKey: privKey,
		pubKey:  privKey.PubKey(),
		addr:    privKey.PubKey().Address(),
	}, nil
}
