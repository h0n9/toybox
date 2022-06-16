package injector

import (
	"github.com/h0n9/toybox/cloud-secrets-injector/cli/injector/template"
	"github.com/spf13/cobra"
)

var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "template related operations",
}

func init() {
	TemplateCmd.AddCommand(
		template.EncodeCmd,
		template.DecodeCmd,
		template.TestCmd,
	)
}
