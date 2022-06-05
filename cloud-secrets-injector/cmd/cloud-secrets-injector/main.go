package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/rs/zerolog"
)

const (
	DefaultProviderName = "aws"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// init context
	ctx := context.Background()

	// get envs
	providerName := util.GetEnv("PROVIDER_NAME", DefaultProviderName)
	secretId := util.GetEnv("SECRET_ID", "")

	if secretId == "" {
		logger.Fatal().Msg("failed to read 'SECRET_ID'")
	}

	switch strings.ToLower(providerName) {
	case "aws":
		providerAWS, err := provider.NewAWS(ctx)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
		secretValue, err := providerAWS.GetSecretValue(secretId)
		if err != nil {
			logger.Err(err)
			break
		}
		fmt.Println(secretValue)
	}
}
