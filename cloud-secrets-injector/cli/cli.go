package cli

import (
	"fmt"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const (
	Name    = "cloud-secrets-injector"
	Version = "v0.0.1"

	DefaultProviderName   = "aws"
	DefaultTemplateBase64 = "e3sgcmFuZ2UgJGssICR2IDo9IC4gfX1be3sgJGsgfX1dCnt7ICR2IH19Cgp7eyBlbmQgfX0K"
	DefaultOutputFilename = "output"
)

var logger *zerolog.Logger

func InitLogger(l *zerolog.Logger) {
	logger = l
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("print '%s' version information", Name),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
