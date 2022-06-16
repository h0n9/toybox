package template

import "github.com/spf13/cobra"

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "test base64-encoded template string",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
