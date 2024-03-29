package p2p

import (
	"context"
	"fmt"
	"os"
	"time"

	coreDiscovery "github.com/libp2p/go-libp2p-core/discovery"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"github.com/postie-labs/go-postie-lib/crypto"

	"github.com/h0n9/toybox/kistio/agent/util"
)

type Node struct {
	ctx context.Context

	privKey *crypto.PrivKey
	pubKey  *crypto.PubKey
	address crypto.Addr

	host             host.Host
	peerDiscovery    *dht.IpfsDHT
	routingDiscovery *discovery.RoutingDiscovery
	backoffDiscovery coreDiscovery.Discovery

	pubSub *pubsub.PubSub
}

func NewNode(ctx context.Context, cfg *util.Config) (*Node, error) {
	privKey, err := crypto.GenPrivKey()
	if err != nil {
		return nil, err
	}
	if cfg.NodePrivKeySeed != "" {
		privKey, err = crypto.GenPrivKeyFromSeed([]byte(cfg.NodePrivKeySeed))
		if err != nil {
			return nil, err
		}
	}

	node := Node{
		ctx:     ctx,
		privKey: privKey,
		pubKey:  privKey.PubKey(),
		address: privKey.PubKey().Address(),
	}

	err = node.NewHost(cfg.NodeListens)
	if err != nil {
		return nil, err
	}

	// init peer discovery alg.
	dhtOptions := []dht.Option{}
	if len(cfg.NodeBootstraps) == 0 || cfg.EnableDHTServer {
		dhtOptions = append(dhtOptions, dht.Mode(dht.ModeServer))
	}
	peerDiscovery, err := dht.New(node.ctx, node.host, dhtOptions...)
	if err != nil {
		return nil, err
	}
	routingDiscovery := discovery.NewRoutingDiscovery(peerDiscovery)
	backoffDiscovery, err := discovery.NewBackoffDiscovery(routingDiscovery, discovery.NewFixedBackoff(1*time.Second))
	if err != nil {
		return nil, err
	}

	node.peerDiscovery = peerDiscovery
	node.routingDiscovery = routingDiscovery
	node.backoffDiscovery = backoffDiscovery

	err = node.NewPubSub()
	if err != nil {
		return nil, err
	}

	return &node, nil
}

func (n *Node) Context() context.Context {
	return n.ctx
}

func (n *Node) Close() error {
	return n.host.Close()
}

func (n *Node) GetHostID() peer.ID {
	return n.host.ID()
}

func (n *Node) GetPeers() []peer.ID {
	return n.host.Network().Peers()
}

func (n *Node) GetPubSub() *pubsub.PubSub {
	return n.pubSub
}

func (n *Node) Info() string {
	if n.host == nil {
		return ""
	}

	str := fmt.Sprintln("host ID:", n.host.ID().Pretty())
	str += fmt.Sprintln("host addrs:", n.host.Addrs())
	str += fmt.Sprintf("%s --bootstrap %s/p2p/%s",
		os.Args[0],
		n.host.Addrs()[0],
		n.host.ID(),
	)

	return str
}
