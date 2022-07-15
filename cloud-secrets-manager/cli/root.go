package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	csm "github.com/h0n9/toybox/cloud-secrets-manager"
	"github.com/h0n9/toybox/cloud-secrets-manager/cli/cert"
	"github.com/h0n9/toybox/cloud-secrets-manager/cli/controller"
	"github.com/h0n9/toybox/cloud-secrets-manager/cli/injector"
	cliTemplate "github.com/h0n9/toybox/cloud-secrets-manager/cli/template"
)

var RootCmd = &cobra.Command{
	Use:   csm.Name,
	Short: fmt.Sprintf("'%s' is a tool for playing with cloud-based secrets", csm.Name),
}

func init() {
	cobra.EnableCommandSorting = false

	RootCmd.AddCommand(
		controller.Cmd,
		injector.Cmd,
		newLineCmd,
		cliTemplate.Cmd,
		cert.Cmd,
		newLineCmd,
		VersionCmd,
	)
}

var newLineCmd = &cobra.Command{Run: func(cmd *cobra.Command, args []string) {}} // new line
