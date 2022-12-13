package server

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "runs msg lake server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
