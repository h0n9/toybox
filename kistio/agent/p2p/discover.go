package p2p

import (
	"fmt"
	"sync"
	"time"

	coreDiscovery "github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"

	"github.com/postie-labs/go-postie-lib/crypto"
)

func (n *Node) connectPeerInfo(pi peer.AddrInfo) error {
	c := n.host.Network().Connectedness(pi.ID)
	if c == network.Connected || c == network.CannotConnect {
		return nil
	}
	err := n.host.Connect(n.ctx, pi)
	if err != nil {
		return err
	}
	fmt.Println("connected:", pi.ID, "peers:", len(n.GetPeers()))
	return nil
}

func (n *Node) connectMultiAddrs(addrs []multiaddr.Multiaddr) error {
	var wg sync.WaitGroup
	for _, addr := range addrs {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = n.connectPeerInfo(*pi)
			if err != nil {
				fmt.Println(err)
				return
			}
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
	return n.connectMultiAddrs(bsNodes)
}

func (n *Node) Advertise(rendezVous string) (time.Duration, error) {
	return n.backoffDiscovery.Advertise(n.ctx, rendezVous, coreDiscovery.TTL(300*time.Millisecond))
}

func (n *Node) Discover(rendezVous string) error {
	peerCh, err := n.backoffDiscovery.FindPeers(n.ctx, rendezVous, coreDiscovery.TTL(300*time.Millisecond))
	if err != nil {
		return err
	}

	for {
		select {
		case <-n.ctx.Done():
			fmt.Println("stop discovering peers")
			return nil
		case peer := <-peerCh:
			if peer.ID == n.GetHostID() || peer.ID == "" {
				continue
			}
			err := n.connectPeerInfo(peer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
