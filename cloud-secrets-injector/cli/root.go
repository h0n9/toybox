package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/h0n9/toybox/cloud-secrets-injector/cli/injector"
	"github.com/h0n9/toybox/cloud-secrets-injector/handler"
	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	DefaultProviderName   = "aws"
	DefaultTemplateBase64 = "e3sgcmFuZ2UgJGssICR2IDo9IC4gfX1be3sgJGsgfX1dCnt7ICR2IH19Cgp7eyBlbmQgfX0K"
	DefaultOutputFilename = "output"
)

var logger *zerolog.Logger

func InitLogger(l *zerolog.Logger) {
	logger = l
}

var RootCmd = &cobra.Command{
	Use:   Name,
	Short: fmt.Sprintf("'%s' is a tool for injecting cloud-based secrets into Docker containers", Name),
	RunE: func(cmd *cobra.Command, args []string) error {
		// init context
		ctx := context.Background()

		logger.Info().Msg("initialized context")

		// get envs
		secretId := util.GetEnv("SECRET_ID", "")
		if secretId == "" {
			return fmt.Errorf("failed to read 'SECRET_ID'")
		}
		providerName := util.GetEnv("PROVIDER_NAME", DefaultProviderName)
		templateBase64 := util.GetEnv("TEMPLATE_BASE64", DefaultTemplateBase64)
		templateFilename := util.GetEnv("TEMPLATE_FILENAME", "")
		outputFilename := util.GetEnv("OUTPUT_FILENAME", DefaultOutputFilename)

		logger.Info().Msg("read environment variables")

		// decode base64-encoded template to string
		templateStr, err := util.DecodeBase64(templateBase64)
		if err != nil {
			return err
		}
		if templateFilename != "" {
			templateStr, err = util.ReadFileToStr(templateFilename)
			if err != nil {
				return err
			}
		}

		logger.Info().Msg("loaded template")

		var (
			secretProvider provider.Provider
			secretHandler  *handler.SecretHandler
		)

		switch strings.ToLower(providerName) {
		case "aws":
			secretProvider, err = provider.NewAWS(ctx)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("failed to figure out secret provider")
		}

		logger.Info().Msg(fmt.Sprintf("initialized secret provider '%s'", providerName))

		secretHandler, err = handler.NewSecretHandler(secretProvider, templateStr)
		if err != nil {
			return err
		}

		logger.Info().Msg("initialized secret handler")

		err = secretHandler.Save(secretId, outputFilename)
		if err != nil {
			return err
		}

		logger.Info().Msg(fmt.Sprintf("saved secret id '%s' values to '%s'", secretId, outputFilename))

		return nil
	},
}

func init() {
	cobra.EnableCommandSorting = false

	RootCmd.AddCommand(
		injector.TemplateCmd,
		VersionCmd,
	)
}
