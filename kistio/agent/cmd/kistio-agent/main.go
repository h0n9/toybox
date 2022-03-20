package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/h0n9/toybox/kistio/agent/p2p"
	"github.com/h0n9/toybox/kistio/agent/util"
)

func main() {
	// prepare
	ctx := context.Background()
	cfg := util.NewConfig()
	cfg.ParseFlags()

	// init node
	node, err := p2p.NewNode(ctx, cfg)
	if err != nil {
		panic(err)
	}

	err = node.DiscoverPeers(cfg.BootstrapNodes)
	if err != nil {
		panic(err)
	}

	fmt.Println(node.Info())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}
