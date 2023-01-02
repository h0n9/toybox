package cli

import (
	"github.com/spf13/cobra"

	"github.com/h0n9/toybox/msg-lake/cli/client"
	"github.com/h0n9/toybox/msg-lake/cli/lab"
	"github.com/h0n9/toybox/msg-lake/cli/server"
)

var RootCmd = &cobra.Command{
	Use:   "lake",
	Short: "simple msg lake",
}

func init() {
	cobra.EnableCommandSorting = false
	RootCmd.AddCommand(
		server.Cmd,
		client.Cmd,
		lab.Cmd,
	)
}
