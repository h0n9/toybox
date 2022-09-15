package p2p

import (
	"bytes"
	"context"
	"fmt"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pDHT "github.com/libp2p/go-libp2p-kad-dht"
	libp2pPubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2pDiscovery "github.com/libp2p/go-libp2p/core/discovery"
	libp2pHost "github.com/libp2p/go-libp2p/core/host"
	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
	discoveryBackoff "github.com/libp2p/go-libp2p/p2p/discovery/backoff"
	discoveryRouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/postie-labs/go-postie-lib/crypto"
	"github.com/rs/zerolog"

	"github.com/h0n9/toybox/kistio/src/util"
)

type Node struct {
	ctx    context.Context
	logger zerolog.Logger

	privKey *crypto.PrivKey
	pubKey  *crypto.PubKey
	addr    crypto.Addr

	host      libp2pHost.Host
	dht       *libp2pDHT.IpfsDHT
	discovery libp2pDiscovery.Discovery

	pubsub *libp2pPubsub.PubSub
	topics map[string]*libp2pPubsub.Topic
}

func NewNode(ctx context.Context, seed []byte, listenAddrs, bootstrapAddrs crypto.Addrs) (*Node, error) {
	// load logger from ctx
	logger, ok := ctx.Value("logger").(zerolog.Logger)
	if !ok {
		return nil, fmt.Errorf("failed to load logger from context")
	}

	// generate private key
	privKey, err := crypto.GenPrivKey()
	if !bytes.Equal(seed, []byte{}) {
		privKey, err = crypto.GenPrivKeyFromSeed(seed)
	}
	if err != nil {
		return nil, err
	}

	// create listenAddr if no received listenAddrs
	if len(listenAddrs) == 0 {
		listenAddr, err := crypto.NewMultiAddr(
			fmt.Sprintf("%s/%d/%s",
				DefaultListenAddr,
				util.GenRandomInt(ListenPortMax, ListenPortMin),
				TransportProtocol,
			),
		)
		if err != nil {
			return nil, err
		}
		listenAddrs = append(listenAddrs, listenAddr)
	}

	// transform privKey to ecdsa based p2p privKey
	privKeyP2P, err := privKey.ToECDSAP2P()
	if err != nil {
		return nil, err
	}

	// init quic transport layer
	quicTransport, err := quic.NewTransport(privKeyP2P, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	// create libp2p host
	host, err := libp2p.New(
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.Identity(privKeyP2P),
		libp2p.Transport(quicTransport),
		libp2p.DefaultSecurity,
	)
	if err != nil {
		return nil, err
	}

	// convert bootstrap multiaddrs to peerinfos
	bootstrapPis, err := libp2pPeer.AddrInfosFromP2pAddrs(bootstrapAddrs...)
	if err != nil {
		return nil, err
	}

	// init dht
	dhtOpts := []libp2pDHT.Option{
		libp2pDHT.BootstrapPeers(bootstrapPis...),
		libp2pDHT.Mode(libp2pDHT.ModeServer),
	}
	dht, err := libp2pDHT.New(ctx, host, dhtOpts...)
	if err != nil {
		return nil, err
	}
	discovery, err := discoveryBackoff.NewBackoffDiscovery(
		discoveryRouting.NewRoutingDiscovery(dht),
		discoveryBackoff.NewFixedBackoff(1*time.Second),
	)
	if err != nil {
		return nil, err
	}

	// init pubsub
	pubsub, err := libp2pPubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, err
	}

	return &Node{
		ctx:    ctx,
		logger: logger,

		privKey: privKey,
		pubKey:  privKey.PubKey(),
		addr:    privKey.PubKey().Address(),

		host:      host,
		dht:       dht,
		discovery: discovery,

		pubsub: pubsub,
		topics: map[string]*libp2pPubsub.Topic{},
	}, nil
}

func (n *Node) Close() {
	for _, topic := range n.topics {
		err := topic.Close()
		if err != nil {
			n.logger.Err(err).Msg("")
		}
	}
	err := n.dht.Close()
	if err != nil {
		n.logger.Err(err).Msg("")
	}
	n.logger.Info().Msg("closed dht")
	err = n.host.Close()
	if err != nil {
		n.logger.Err(err).Msg("")
	}
	n.logger.Info().Msg("closed host")
}
