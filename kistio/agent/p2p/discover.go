package p2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/multiformats/go-multiaddr"

	"github.com/postie-labs/go-postie-lib/crypto"
)

func (n *Node) connectPeerInfo(pi peer.AddrInfo) error {
	if n.host.Network().Connectedness(pi.ID) == network.Connected {
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

func (n *Node) Discover(rendezVous string) error {
	// advertise rendez-vous annoucement
	routingDiscovery := discovery.NewRoutingDiscovery(n.peerDiscovery)
	discovery.Advertise(n.ctx, routingDiscovery, rendezVous)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			fmt.Println("stop discovering peers")
			return nil
		case <-ticker.C:
			pis, err := discovery.FindPeers(n.ctx, routingDiscovery, rendezVous)
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, pi := range pis {
				if pi.ID == n.GetHostID() {
					continue
				}
				err := n.connectPeerInfo(pi)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
}
