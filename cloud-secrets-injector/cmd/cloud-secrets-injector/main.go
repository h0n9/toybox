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
	DefaultProviderName   = "aws"
	DefaultTemplateBase64 = "e3sgcmFuZ2UgJGssICR2IDo9IC4gfX1be3sgJGsgfX1dCnt7ICR2IH19Cgp7eyBlbmQgfX0K"
	DefaultOutputFilename = "output"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	// init context
	ctx := context.Background()

	// get envs
	providerName := util.GetEnv("PROVIDER_NAME", DefaultProviderName)
	secretId := util.GetEnv("SECRET_ID", "")
	templateBase64 := util.GetEnv("TEMPLATE_BASE64", DefaultTemplateBase64)
	outputFilename := util.GetEnv("OUTPUT_FILENAME", DefaultOutputFilename)

	if secretId == "" {
		logger.Fatal().Msg("failed to read 'SECRET_ID'")
	}

	// decode base64-encoded template to string
	templateStr, err := util.DecodeBase64(templateBase64)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	var secretHandler *handler.SecretHandler

	switch strings.ToLower(providerName) {
	case "aws":
		providerAWS, err := provider.NewAWS(ctx)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
		secretHandler, err = handler.NewSecretHandler(providerAWS, templateStr)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
	default:
		logger.Fatal().Msg("failed to figure out the provider")
	}

	err = secretHandler.Save(secretId, outputFilename)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
