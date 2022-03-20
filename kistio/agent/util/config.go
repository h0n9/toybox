package util

import (
	"flag"

	"github.com/postie-labs/go-postie-lib/crypto"
)

type Config struct {
	BootstrapNodes crypto.Addrs
	ListenAddrs    crypto.Addrs
	PrivKeySeed    string
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) ParseFlags() error {
	flag.Var(&cfg.ListenAddrs, "listen", "addresses to listen from")
	flag.Var(&cfg.BootstrapNodes, "bootstrap", "bootstrap nodes")
	privKeySeed := flag.String("seed", "", "private key seed")

	flag.Parse()

	cfg.PrivKeySeed = *privKeySeed

	return nil
}
