package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
)

const (
	Name = "standard-main"
)

func main() {
	// init logger
	logger := zerolog.New(os.Stderr).With().
		Timestamp().
		Str("service", Name).
		Logger()

	// init context
	_, cancel := context.WithCancel(context.Background())

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

	// wait until all of wait groups are done
	wg.Wait()
}
