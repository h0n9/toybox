package p2p

import (
	"math/rand"
)

// Multi Address used on QUIC protocol is formed as follows:
// ex) /ip4/0.0.0.0/udp/61881/quic

const (
	TransportProtocol = "quic"
	ProtocolID        = "/kistio-agent/1.0.0"

	DefaultListenAddr = "/ip4/0.0.0.0/udp"
	MinListenPort     = 49152
	MaxListenPort     = 65535
)

func genRandPortNum() int {
	return rand.Intn(MaxListenPort-MinListenPort) + MinListenPort
}
