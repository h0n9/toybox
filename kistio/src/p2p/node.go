package p2p

import (
	"bytes"
	"context"
	"fmt"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pDiscovery "github.com/libp2p/go-libp2p-core/discovery"
	libp2pHost "github.com/libp2p/go-libp2p-core/host"
	libp2pDHT "github.com/libp2p/go-libp2p-kad-dht"
	discoveryBackoff "github.com/libp2p/go-libp2p/p2p/discovery/backoff"
	discoveryRouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/postie-labs/go-postie-lib/crypto"

	"github.com/h0n9/toybox/kistio/src/util"
)

type Node struct {
	ctx    context.Context
	logger zerolog.Logger

	privKey *crypto.PrivKey
	pubKey  *crypto.PubKey
	addr    crypto.Addr

	host      libp2pHost.Host
	discovery libp2pDiscovery.Discovery
}

func NewNode(ctx context.Context, seed []byte, listenAddrs crypto.Addrs, dhtModeServer bool) (*Node, error) {
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

	// init dht
	dhtOpts := []libp2pDHT.Option{}
	if dhtModeServer {
		dhtOpts = append(dhtOpts, libp2pDHT.Mode(libp2pDHT.ModeServer))
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

	return &Node{
		ctx:    ctx,
		logger: logger,

		privKey: privKey,
		pubKey:  privKey.PubKey(),
		addr:    privKey.PubKey().Address(),

		host:      host,
		discovery: discovery,
	}, nil
}

func (n *Node) Bootstrap(addrs crypto.Addrs) error {
	return nil
}
