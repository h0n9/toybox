package template

import "github.com/spf13/cobra"

var EncodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "encode human-readable template string to base64-encoded string",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
