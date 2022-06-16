package cli

import (
	"github.com/h0n9/toybox/cloud-secrets-injector/cli/injector"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cloud-secrets-injector",
	Short: "'cloud-secrets-injector' is a tool for injecting cloud-based secrets into Docker containers",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	cobra.EnableCommandSorting = false

	RootCmd.AddCommand(
		injector.TemplateCmd,
	)
}
