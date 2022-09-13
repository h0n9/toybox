package p2p

import (
	"fmt"
	"sync"

	libp2pPeer "github.com/libp2p/go-libp2p/core/peer"
)

func (n *Node) Connect(pi libp2pPeer.AddrInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	err := n.host.Connect(n.ctx, pi)
	if err != nil {
		n.logger.Err(err).Msg("")
		return
	}
	n.logger.Info().Msgf("connected to %s", pi.ID)
}

// getter, setter
func (n *Node) GetHostID() libp2pPeer.ID {
	return n.host.ID()
}

func (n *Node) GetAddr() string {
	addrs := n.host.Addrs()
	if len(addrs) < 1 {
		return ""
	}
	return fmt.Sprintf("%s/p2p/%s", addrs[0], n.host.ID())
}
