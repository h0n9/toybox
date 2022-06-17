package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Name    = "cloud-secrets-injector"
	Version = "v0.0.1"
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("print '%s' version information", Name),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}
