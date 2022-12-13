package client

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "client",
	Short: "runs msg lake client (interactive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
