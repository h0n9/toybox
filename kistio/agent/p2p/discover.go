package p2p

import (
	"fmt"
	"sync"

	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"

	"github.com/postie-labs/go-postie-lib/crypto"
)

func (n *Node) DiscoverPeers(bsNodes crypto.Addrs) error {
	// init peer discovery alg.
	peerDiscovery, err := dht.New(n.ctx, n.host)
	if err != nil {
		return err
	}

	// bootstrap peer discovery
	err = peerDiscovery.Bootstrap(n.ctx)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for _, bsn := range bsNodes {
		peerInfo, err := peer.AddrInfoFromP2pAddr(bsn)
		if err != nil {
			panic(err)
		}

		wg.Add(1)

		go func() {
			defer wg.Done()
			err = n.host.Connect(n.ctx, *peerInfo)
			if err != nil {
				panic(err)
			}

			fmt.Println("connected to:", *peerInfo)
		}()

	}
	wg.Wait()

	// advertise rendez-vous annoucement
	routingDiscovery := discovery.NewRoutingDiscovery(peerDiscovery)
	discovery.Advertise(n.ctx, routingDiscovery, RendezVous)

	peers, err := routingDiscovery.FindPeers(n.ctx, RendezVous)
	if err != nil {
		return err
	}

	for peer := range peers {
		if peer.ID == n.host.ID() {
			continue
		}

		// stream, err := n.host.NewStream(n.ctx, peer.ID, ProtocolID)
		err = n.host.Connect(n.ctx, peer)
		if err != nil {
			fmt.Println("failed to connect to:", peer)
			continue
		}

		fmt.Println("connected to:", peer)
		// handleStream(stream)
	}

	return nil
}
