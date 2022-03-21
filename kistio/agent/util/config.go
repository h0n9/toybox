package util

import (
	"flag"

	"github.com/postie-labs/go-postie-lib/crypto"
)

type Config struct {
	NodeListens     crypto.Addrs
	NodeBootstraps  crypto.Addrs
	NodePrivKeySeed string

	GrpcListen string
}

const (
	DefaultGrpcListen = "0.0.0.0:7788"
)

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) ParseFlags() error {
	flag.Var(&cfg.NodeListens, "listens", "addresses to listen from")
	flag.Var(&cfg.NodeBootstraps, "bootstraps", "bootstrap nodes")
	privKeySeed := flag.String("seed", "", "private key seed")
	grpcListen := flag.String("grpc-listen", DefaultGrpcListen, "grpc listen host:port")

	flag.Parse()

	cfg.NodePrivKeySeed = *privKeySeed
	cfg.GrpcListen = *grpcListen

	return nil
}
