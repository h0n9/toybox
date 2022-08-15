package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	kistio "github.com/h0n9/toybox/kistio/src"
)

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "connector node",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a server for connector node",
	Run: func(cmd *cobra.Command, args []string) {
		logger := zerolog.New(os.Stderr).With().
			Timestamp().
			Str("service", fmt.Sprintf("%s-%s", kistio.Name, Cmd.Use)).
			Logger()
		_, cancel := context.WithCancel(context.Background())
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		wg := sync.WaitGroup{}

		logger.Info().Msg("initialized logger, context, sig channel, wait group")

		wg.Add(1)
		go func() {
			defer wg.Done()
			sig := <-sigCh
			logger.Info().Msg("receieved signal " + sig.String())
			cancel()
		}()

		wg.Wait()
	},
}

func init() {
	Cmd.AddCommand(runCmd)
}
