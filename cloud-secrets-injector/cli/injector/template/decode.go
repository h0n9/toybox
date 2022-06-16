package template

import "github.com/spf13/cobra"

var DecodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "decode base64-encoded template string to human-readable string",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
