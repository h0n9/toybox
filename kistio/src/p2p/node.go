package p2p

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	libp2p "github.com/libp2p/go-libp2p"
	libp2pDiscovery "github.com/libp2p/go-libp2p-core/discovery"
	libp2pHost "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	libp2pDHT "github.com/libp2p/go-libp2p-kad-dht"
	discoveryBackoff "github.com/libp2p/go-libp2p/p2p/discovery/backoff"
	discoveryRouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	quic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/multiformats/go-multiaddr"
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
		dht:       dht,
		discovery: discovery,
	}, nil
}

func (n *Node) Bootstrap(addrs crypto.Addrs) error {
	err := n.dht.Bootstrap(n.ctx)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, addr := range addrs {
		wg.Add(1)
		go func(addr multiaddr.Multiaddr) {
			defer wg.Done()
			pi, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				n.logger.Err(err)
				return
			}
			n.host.Connect(n.ctx, *pi)
		}(addr)
	}
	wg.Wait()
	return nil
}

func (n *Node) Discover(rendezVous string) error {
	peerCh, err := n.discovery.FindPeers(
		n.ctx,
		rendezVous,
		libp2pDiscovery.Limit(3),
		libp2pDiscovery.TTL(100*time.Millisecond),
	)
	if err != nil {
		return err
	}
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
				_, err := n.discovery.Advertise(n.ctx, rendezVous)
				if err != nil {
					n.logger.Err(err)
				}
			}
		}
	}()

	// find peers
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-n.ctx.Done():
				n.logger.Info().Msg("stop finding peers")
				return
			case pi := <-peerCh:
				if pi.ID == "" || pi.ID == n.host.ID() {
					continue
				}
				err = n.host.Connect(n.ctx, pi)
				if err != nil {
					n.logger.Err(err)
					continue
				}
				n.logger.Info().Msgf("connected to %s", pi.ID)
			}
		}
	}()

	wg.Wait()

	return nil
}
