package template

import (
	"fmt"

	"github.com/h0n9/toybox/cloud-secrets-injector/util"
	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "encode human-readable template string to base64-encoded string",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if args[0] == "" {
			return fmt.Errorf("failed to encode empty string")
		}
		fmt.Println(util.EncodeBase64(args[0]))
		return nil
	},
}
