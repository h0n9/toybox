package p2p

import (
	"fmt"
	"sync"
	"time"

	coreDiscovery "github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
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

func (n *Node) Discover(rendezVous string) error {
	// advertise rendez-vous annoucement
	routingDiscovery := discovery.NewRoutingDiscovery(n.peerDiscovery)
	backoffDiscovery, err := discovery.NewBackoffDiscovery(routingDiscovery, discovery.NewFixedBackoff(1*time.Second))
	if err != nil {
		return err
	}

	opts := []coreDiscovery.Option{coreDiscovery.TTL(300 * time.Millisecond)}

	go backoffDiscovery.Advertise(n.ctx, rendezVous, opts...)

	go func() {
		peerCh, err := backoffDiscovery.FindPeers(n.ctx, rendezVous, opts...)
		if err != nil {
			fmt.Println(err)
			return
		}

		for {
			select {
			case <-n.ctx.Done():
				fmt.Println("stop discovering peers")
				return
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
	}()

	return nil
}
