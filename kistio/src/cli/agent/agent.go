package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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
	Use:   "agent",
	Short: "connector node",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "run a server for connector node",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			node *p2p.Node
		)

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
			if node != nil {
				node.Close()
			}
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
		node, err := p2p.NewNode(ctx, seedBytes, listenAddrs, bootstrapAddrs)
		if err != nil {
			logger.Err(err).Msg("")
			return
		}
		logger.Info().Msgf("initialized node: %s", node.GetAddr())

		// bootstrap node
		err = node.Bootstrap()
		if err != nil {
			logger.Err(err).Msg("")
			return
		}
		logger.Info().Msg("bootstrapped peer nodes")

		// discover peers
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := node.Discover(rendezVous)
			if err != nil {
				logger.Err(err).Msg("")
				return
			}
		}()

		// say something
		topic, err := node.JoinTopic("hello world")
		if err != nil {
			logger.Err(err).Msg("")
			return
		}
		sub, err := topic.Subscribe()
		if err != nil {
			logger.Err(err).Msg("")
			return
		}
		defer sub.Cancel()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				msg, err := sub.Next(ctx)
				if err != nil {
					logger.Err(err).Msg("")
					break
				}
				if msg.GetFrom() == node.GetHostID() {
					continue
				}
				logger.Info().Msgf("%s - %s (%d)", msg.GetData(), msg.GetFrom(), len(topic.ListPeers()))
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			ticker := time.NewTicker(1 * time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					err = topic.Publish(ctx, []byte("hello world"))
					if err != nil {
						logger.Err(err).Msg("")
					}
				}
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
