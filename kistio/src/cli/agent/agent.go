package agent

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "kistio-agent: connector node",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a server for connector node",
}

func init() {
	Cmd.AddCommand(runCmd)
}
