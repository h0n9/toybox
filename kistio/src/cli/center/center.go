package center

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "center",
	Short: "seed node, admission webhook controller",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a server for seed node and managing admission webhooks",
}

func init() {
	Cmd.AddCommand(runCmd)
}
