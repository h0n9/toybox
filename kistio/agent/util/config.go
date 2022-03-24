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

	RendezVous      string
	EnableDHTServer bool
}

const (
	DefaultGrpcListen      = "0.0.0.0:7788"
	DefaultRendezVous      = "t'as bien dormi ?"
	DefaultEnableDHTServer = false
)

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) ParseFlags() error {
	flag.Var(&cfg.NodeListens, "listens", "addresses to listen from")
	flag.Var(&cfg.NodeBootstraps, "bootstraps", "bootstrap nodes")
	privKeySeed := flag.String("seed", "", "private key seed")
	grpcListen := flag.String("grpc-listen", DefaultGrpcListen, "grpc listen host:port")
	rendezVous := flag.String("rendez-vous", DefaultRendezVous, "rendez-vous point for node discovery")
	enableDHTModeServer := flag.Bool("enable-dht-server", DefaultEnableDHTServer, "enable dht server mode")

	flag.Parse()

	cfg.NodePrivKeySeed = *privKeySeed
	cfg.GrpcListen = *grpcListen
	cfg.RendezVous = *rendezVous
	cfg.EnableDHTServer = *enableDHTModeServer

	return nil
}
