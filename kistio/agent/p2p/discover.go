package p2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/multiformats/go-multiaddr"

	"github.com/postie-labs/go-postie-lib/crypto"
)

func (n *Node) connect(addrs []multiaddr.Multiaddr) error {
	var wg sync.WaitGroup
	for _, addr := range addrs {
		peerInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			err = n.host.Connect(n.ctx, *peerInfo)
			if err != nil {
				panic(err)
			}
			fmt.Println("peers:", n.GetPeers())
		}()
	}
	wg.Wait()
	return nil
}

func (n *Node) Bootstrap(bsNodes crypto.Addrs) error {
	// bootstrap peer discovery
	err := n.peerDiscovery.Bootstrap(n.ctx)
	if err != nil {
		return err
	}
	return n.connect(bsNodes)
}

func (n *Node) DiscoverMDNS() error {
	return nil
}

func (n *Node) DiscoverDHT() error {
	// advertise rendez-vous annoucement
	routingDiscovery := discovery.NewRoutingDiscovery(n.peerDiscovery)
	discovery.Advertise(n.ctx, routingDiscovery, RendezVous)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return nil
		case <-ticker.C:
			peers, err := discovery.FindPeers(n.ctx, routingDiscovery, RendezVous)
			if err != nil {
				return err
			}
			for _, p := range peers {
				if p.ID == n.GetHostID() {
					continue
				}

				err := n.connect(p.Addrs)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
