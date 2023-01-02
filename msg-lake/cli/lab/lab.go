package lab

import (
	"github.com/spf13/cobra"

	"github.com/h0n9/toybox/msg-lake/cli/lab/client"
)

var Cmd = &cobra.Command{
	Use:   "lab",
	Short: "laboratory 👨‍🔬",
}

func init() {
	cobra.EnableCommandSorting = false
	Cmd.AddCommand(
		client.Cmd,
	)
}
