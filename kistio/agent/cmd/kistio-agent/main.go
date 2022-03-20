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

	tp, err := node.Join("test")
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}

	// publish
	wg.Add(1)
	go func() {
		for {
			// publish simple data for test
			err = tp.Publish(ctx, []byte(fmt.Sprintf("%s - hello world", time.Now().String())))
			time.Sleep(1 * time.Second)
		}
	}()

	// subscribe
	wg.Add(1)
	go func() {
		sub, err := tp.Subscribe()
		if err != nil {
			panic(err)
		}
		for {
			msg, err := sub.Next(ctx)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if msg.GetFrom() == node.GetHostID() {
				continue
			}
			fmt.Printf("[%s] %s\n", msg.GetFrom().Pretty(), msg.GetData())
		}
	}()

	wg.Wait()
}
