package center

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/postie-labs/go-postie-lib/crypto"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	kistio "github.com/h0n9/toybox/kistio/src"
	"github.com/h0n9/toybox/kistio/src/p2p"
)

var (
	seed           string
	listenAddrs    crypto.Addrs
	bootstrapAddrs crypto.Addrs
	rendezVous     string
)

var Cmd = &cobra.Command{
	Use:   "center",
	Short: "seed node, admission webhook controller",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a server for seed node and managing admission webhooks",
	Run: func(cmd *cobra.Command, args []string) {
		// init logger
		logger := zerolog.New(os.Stderr).With().
			Timestamp().
			Str("service", fmt.Sprintf("%s-%s", kistio.Name, Cmd.Use)).
			Logger()

		// init context
		ctx, cancel := context.WithCancel(context.Background())
		ctx = context.WithValue(ctx, "logger", logger)

		// init sig channel
		sigCh := make(chan os.Signal, 1)
		defer close(sigCh)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// init wait group
		wg := sync.WaitGroup{}

		logger.Info().Msg("initialized logger, context, sig channel, wait group")

		// init goroutine for watching signals
		wg.Add(1)
		go func() {
			defer wg.Done()
			sig := <-sigCh
			logger.Info().Msg("receieved signal " + sig.String())
			cancel()
		}()

		// **********************
		// * do something here *
		// **********************

		// transform seed string to byte slice if not ""
		seedBytes := []byte{}
		if seed != "" {
			seedBytes = []byte(seed)
		}

		// init node
		node, err := p2p.NewNode(ctx, seedBytes, listenAddrs, true)
		if err != nil {
			logger.Err(err)
			return
		}
		logger.Info().Msgf("initialized node: %s", node.GetAddr())

		// bootstrap node
		err = node.Bootstrap(bootstrapAddrs...)
		if err != nil {
			logger.Err(err)
			return
		}
		logger.Info().Msg("bootstrapped peer nodes")

		// discover peers
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := node.Discover(rendezVous)
			if err != nil {
				logger.Err(err)
				return
			}
		}()

		// wait until all of wait groups are done
		wg.Wait()
	},
}

func init() {
	runCmd.Flags().StringVarP(&seed, "seed", "s", "", "seed for private key")
	runCmd.Flags().VarP(&listenAddrs, "listen", "l", "listening addresses")
	runCmd.Flags().VarP(&bootstrapAddrs, "bootstrap", "b", "bootstrap address")
	runCmd.Flags().StringVar(&rendezVous, "rendez-vous", "", "rendez-vous point")
	Cmd.AddCommand(runCmd)
}
