package relayer

import (
	"bufio"
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"

	"github.com/multiformats/go-multiaddr"
)

const (
	protocolID      = protocol.ID("/msg-lake/v1.0-beta-0")
	mdnsServiceName = "_p2p_msg-lake._udp"
)

type Relayer struct {
	ctx context.Context

	privKey crypto.PrivKey
	pubKey  crypto.PubKey

	h host.Host

	peerChan <-chan peer.AddrInfo
}

func NewRelayer(ctx context.Context, hostname string, port int) (*Relayer, error) {
	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return nil, err
	}

	ma, err := multiaddr.NewMultiaddr(
		fmt.Sprintf(
			"/ip4/%s/udp/%d/quic",
			hostname,
			port,
		),
	)
	if err != nil {
		return nil, err
	}

	h, err := libp2p.New(
		libp2p.ListenAddrs(ma),
		libp2p.Identity(privKey),
		libp2p.Transport(libp2pquic.NewTransport),
	)
	if err != nil {
		return nil, err
	}

	// register stream handler
	h.SetStreamHandler(protocolID, handleStream)

	// init mdns service
	dn := newDiscoveryNotifee()
	svc := mdns.NewMdnsService(h, mdnsServiceName, dn)
	err = svc.Start()
	if err != nil {
		return nil, err
	}

	fmt.Printf("listening on: %s\n", ma)

	return &Relayer{
		ctx: ctx,

		privKey: privKey,
		pubKey:  pubKey,

		h: h,

		peerChan: dn.peerChan,
	}, nil
}

func (relayer *Relayer) DiscoverPeers() error {
	for {
		fmt.Println("waiting peers ...")
		peer := <-relayer.peerChan // blocks until discover new peers
		fmt.Printf("found peer: %s\n", peer)

		fmt.Printf("connecting peer: %s\n", peer)
		err := relayer.h.Connect(relayer.ctx, peer)
		if err != nil {
			fmt.Printf("failed to connect peer: %s\n", peer)
			continue
		}

		stream, err := relayer.h.NewStream(relayer.ctx, peer.ID, protocolID)
		if err != nil {
			fmt.Println(err)
			continue
		}
		rw := bufio.NewReadWriter(
			bufio.NewReader(stream),
			bufio.NewWriter(stream),
		)
		go writeData(relayer.h.ID().String(), rw)

		fmt.Printf("connected to peer: %s\n", peer)
	}
}

func handleStream(s network.Stream) {
	fmt.Println("got a new stream")
	rw := bufio.NewReadWriter(
		bufio.NewReader(s),
		bufio.NewWriter(s),
	)
	go readData(s.ID(), rw)
}

func readData(id string, rw *bufio.ReadWriter) {
	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from buffer")
			break
		}

		if str == "" {
			return
		}
		if str != "\n" {
			// Green console colour: 	\x1b[32m
			// Reset console colour: 	\x1b[0m
			fmt.Printf("\x1b[32m%s\x1b[0m> ", str)
		}
	}
}

func writeData(id string, rw *bufio.ReadWriter) {
	for {
		data := fmt.Sprintf("%s - %s\n", id, time.Now().String())
		_, err := rw.WriteString(data)
		if err != nil {
			fmt.Println("Error writing to buffer")
			break
		}
		err = rw.Flush()
		if err != nil {
			fmt.Println("Error flushing buffer")
			break
		}
		time.Sleep(1 * time.Second)
	}
}

type discoveryNotifee struct {
	peerChan chan peer.AddrInfo
}

func newDiscoveryNotifee() *discoveryNotifee {
	return &discoveryNotifee{
		peerChan: make(chan peer.AddrInfo),
	}
}

// interface to be called when new  peer is found
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.peerChan <- pi
}
