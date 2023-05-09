package main

import (
	"context"
	"math/rand"

	"github.com/h0n9/toybox/msg-lake/relayer"
)

func main() {
	ctx := context.Background()
	port := rand.Intn(8080-1000+1) + 1000
	relayer, err := relayer.NewRelayer(ctx, "127.0.0.1", port)
	if err != nil {
		panic(err)
	}
	relayer.DiscoverPeers()
}
