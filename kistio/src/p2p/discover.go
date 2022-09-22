package p2p

import (
	"sync"
	"time"

	libp2pDiscovery "github.com/libp2p/go-libp2p/core/discovery"
)

func (n *Node) Bootstrap() error {
	return n.dht.Bootstrap(n.ctx)
}

func (n *Node) Discover(rendezVous string) error {
	ticker := time.NewTicker(1000 * time.Millisecond)
	wg := sync.WaitGroup{}

	// advertise
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ticker.Stop()
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info().Msg("stop advertising")
				return
			case <-ticker.C:
				// skip advertising when node has no peers in routing table
				routingTableSize := n.dht.RoutingTable().Size()

				n.logger.Debug().Msgf("peers: %v", n.dht.RoutingTable().ListPeers())
				n.logger.Debug().Msgf("routing table size: %d", routingTableSize)
				if routingTableSize < 1 {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					n.logger.Debug().Msg("advertising")
					_, err := n.discovery.Advertise(n.ctx, rendezVous)
					if err != nil {
						n.logger.Err(err).Msg("")
					}
				}()
			}
		}
	}()

	// find peers
	wg.Add(1)
	go func() {
		defer wg.Done()
		peerCh, err := n.discovery.FindPeers(
			n.ctx,
			rendezVous,
			libp2pDiscovery.Limit(3),
			libp2pDiscovery.TTL(100*time.Millisecond),
		)
		if err != nil {
			n.logger.Fatal().Err(err)
		}
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info().Msg("stop finding peers")
				return
			case pi := <-peerCh:
				if pi.ID == "" || pi.ID == n.host.ID() {
					continue
				}
				wg.Add(1)
				go n.Connect(pi, &wg)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()

	return nil
}
