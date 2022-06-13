package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/h0n9/toybox/cloud-secrets-injector/handler"
	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/rs/zerolog"
)

const (
	DefaultProviderName   = "aws"
	DefaultTemplateBase64 = "e3sgcmFuZ2UgJGssICR2IDo9IC4gfX1be3sgJGsgfX1dCnt7ICR2IH19Cgp7eyBlbmQgfX0K"
	DefaultOutputFilename = "output"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// init context
	ctx := context.Background()

	logger.Info().Msg("initialized context")

	// get envs
	secretId := util.GetEnv("SECRET_ID", "")
	if secretId == "" {
		logger.Fatal().Msg("failed to read 'SECRET_ID'")
	}
	providerName := util.GetEnv("PROVIDER_NAME", DefaultProviderName)
	templateBase64 := util.GetEnv("TEMPLATE_BASE64", DefaultTemplateBase64)
	outputFilename := util.GetEnv("OUTPUT_FILENAME", DefaultOutputFilename)

	logger.Info().Msg("read environment variables")

	// decode base64-encoded template to string
	templateStr, err := util.DecodeBase64(templateBase64)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	logger.Info().Msg("decoded base64-encoded template to string")

	var (
		secretProvider provider.Provider
		secretHandler  *handler.SecretHandler
	)

	switch strings.ToLower(providerName) {
	case "aws":
		secretProvider, err = provider.NewAWS(ctx)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
	default:
		logger.Fatal().Msg("failed to figure out secret provider")
	}

	logger.Info().Msg(fmt.Sprintf("initialized secret provider '%s'", providerName))

	secretHandler, err = handler.NewSecretHandler(secretProvider, templateStr)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	logger.Info().Msg("initialized secret handler")

	err = secretHandler.Save(secretId, outputFilename)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	logger.Info().Msg(fmt.Sprintf("saved secret id '%s' values to '%s'", secretId, outputFilename))
}
