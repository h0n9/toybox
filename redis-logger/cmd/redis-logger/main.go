package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/h0n9/toybox/redis-logger/util"
	"github.com/rs/zerolog/log"
)

const (
	DefaultServiceName   = "redis-logger"
	DefaultRedisAddr     = "localhost:6379"
	DefaultRedisUsername = ""
	DefaultRedisPassword = ""
)

func main() {
	// get envs
	serviceName := util.GetEnv("SERVICE_NAME", DefaultServiceName)
	redisAddr := util.GetEnv("REDIS_ADDR", DefaultRedisAddr)
	redisUsername := util.GetEnv("REDIS_USERNAME", DefaultRedisUsername)
	redisPassword := util.GetEnv("REDIS_PASSWORD", DefaultRedisPassword)
	redisEnableTLS := util.GetEnv("REDIS_ENABLE_TLS", "")

	// init context, waitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	// init signal handler
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// init logger
	logger := log.With().
		Str("service", serviceName).
		Str("redis-addr", redisAddr).
		Logger()

	// wait signals
	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := <-sigs // block until signal
		logger.Info().Msg(fmt.Sprintf("received SIGNAL: %s", sig.String()))
		cancel()
	}()

	// init redis client
	redisOpts := redis.Options{
		Addr:     redisAddr,
		Username: redisUsername,
		Password: redisPassword,
		DB:       0,
	}
	if redisEnableTLS != "" {
		redisOpts.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	rdb := redis.NewClient(&redisOpts)
	defer rdb.Close()

	// test connection
	ping := rdb.Ping(ctx)
	result, err := ping.Result()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}
	logger.Info().Msg(result)

	// get slowlogs
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(1 * time.Second)

		for {
			select {
			case <-ctx.Done():
				logger.Info().Msg("stop getting slowlogs")
				return
			case <-tick:
			}
			cmd := redis.NewSlowLogCmd(ctx, "slowlog", "get", 1)
			rdb.Process(ctx, cmd)
			slowlogs, err := cmd.Result()
			if err != nil {
				logger.Err(err)
				continue
			}
			for _, slowlog := range slowlogs {
				if strings.HasPrefix(slowlog.Args[0], "slowlog") {
					continue
				}
				logger.Info().
					Str("type", "SLOWLOG").
					Str("client-addr", slowlog.ClientAddr).
					Str("client-name", slowlog.ClientName).
					Str("duration", slowlog.Duration.String()).
					Msg(fmt.Sprint(slowlog.Args))
			}
		}
	}()

	wg.Wait()
}
