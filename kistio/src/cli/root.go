package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	kistio "github.com/h0n9/toybox/kistio/src"
)

var RootCmd = &cobra.Command{
	Use:   kistio.Name,
	Short: fmt.Sprintf("'%s' is a lightweight infrastructure platform for sharing messages between services", kistio.Name),
}

func init() {
	cobra.EnableCommandSorting = false

	RootCmd.AddCommand()
}
