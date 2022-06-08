package main

import (
	"context"
	"os"
	"strings"

	"github.com/h0n9/toybox/cloud-secrets-injector/handler"
	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/rs/zerolog"
)

const (
	DefaultProviderName     = "aws"
	DefaultTemplateFilename = "sample-template"
	DefaultOutputFilename   = "output"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// init context
	ctx := context.Background()

	// get envs
	providerName := util.GetEnv("PROVIDER_NAME", DefaultProviderName)
	secretId := util.GetEnv("SECRET_ID", "")
	templateFilename := util.GetEnv("TEMPLATE_FILENAME", DefaultTemplateFilename)
	outputFilename := util.GetEnv("OUTPUT_FILENAME", DefaultOutputFilename)

	if secretId == "" {
		logger.Fatal().Msg("failed to read 'SECRET_ID'")
	}

	var secretHandler *handler.SecretHandler

	switch strings.ToLower(providerName) {
	case "aws":
		providerAWS, err := provider.NewAWS(ctx)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
		secretHandler, err = handler.NewSecretHandler(providerAWS, templateFilename)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
	default:
		logger.Fatal().Msg("failed to figure out the provider")
	}

	err := secretHandler.Save(secretId, outputFilename)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
