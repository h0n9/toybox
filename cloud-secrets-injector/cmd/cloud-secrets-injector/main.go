package main

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"text/template"

	"github.com/h0n9/toybox/cloud-secrets-injector/handler"
	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/rs/zerolog"
)

const (
	DefaultProviderName = "aws"
	SampleTemplate      = "{{ range $k, $v := . }}export {{ $k }}={{ $v }}\n{{ end }}"
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

	var secretHandler *handler.SecretHandler

	switch strings.ToLower(providerName) {
	case "aws":
		providerAWS, err := provider.NewAWS(ctx)
		if err != nil {
			logger.Fatal().Msg(err.Error())
		}
		secretHandler = handler.NewSecretHandler(providerAWS)
	default:
		logger.Fatal().Msg("failed to figure out the provider")
	}

	secretValue, err := secretHandler.Get(secretId)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	var m map[string]interface{}

	err = json.Unmarshal([]byte(secretValue), &m)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	tmpl := template.New("sample-template")
	tmpl, err = tmpl.Parse(SampleTemplate)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}

	err = tmpl.Execute(os.Stdout, m)
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
}
