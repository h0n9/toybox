package main

import (
	"os"

	"github.com/h0n9/toybox/cloud-secrets-injector/cli"
	"github.com/rs/zerolog"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	cli.InitLogger(&logger)

	err := cli.RootCmd.Execute()
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
