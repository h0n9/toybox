package p2p

import (
	"context"
	"fmt"
	"os"

	"github.com/h0n9/toybox/kistio/agent/util"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/postie-labs/go-postie-lib/crypto"
)

type Node struct {
	ctx context.Context

	privKey *crypto.PrivKey
	pubKey  *crypto.PubKey
	address crypto.Addr

	host          host.Host
	peerDiscovery *dht.IpfsDHT

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
	if len(cfg.NodeBootstraps) == 0 {
		dhtOptions = append(dhtOptions, dht.Mode(dht.ModeServer))
	}
	peerDiscovery, err := dht.New(node.ctx, node.host, dhtOptions...)
	if err != nil {
		return nil, err
	}
	node.peerDiscovery = peerDiscovery

	err = node.NewPubSub()
	if err != nil {
		return nil, err
	}

	return &node, nil
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
