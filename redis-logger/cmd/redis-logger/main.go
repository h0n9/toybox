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
	DefaultTimeInterval  = "1000ms"
)

var CmdsToIgnore = map[string]bool{
	"slowlog": true,
	"client":  true,
	"auth":    true,
	"ping":    true,
}

func main() {
	// get envs
	serviceName := util.GetEnv("SERVICE_NAME", DefaultServiceName)
	redisAddr := util.GetEnv("REDIS_ADDR", DefaultRedisAddr)
	redisUsername := util.GetEnv("REDIS_USERNAME", DefaultRedisUsername)
	redisPassword := util.GetEnv("REDIS_PASSWORD", DefaultRedisPassword)
	redisEnableTLS := util.GetEnv("REDIS_ENABLE_TLS", "")
	timeIntervalStr := util.GetEnv("TIME_INTERVAL", DefaultTimeInterval)

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

	// convert time interval string to int64
	timeInterval, err := time.ParseDuration(timeIntervalStr)
	if err != nil {
		logger.Panic().Msg(err.Error())
	}

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

	// global variables
	clientList := map[string]string{} // <client-ip>:<client-username>

	// get slowlogs
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(timeInterval)

		for {
			select {
			case <-ctx.Done():
				logger.Info().Msg("stop getting slowlogs")
				return
			case <-tick:
				go func() {
					cmd := redis.NewSlowLogCmd(ctx, "slowlog", "get", 1)
					err = rdb.Process(ctx, cmd)
					if err != nil {
						logger.Err(err)
						return
					}
					slowlogs, err := cmd.Result()
					if err != nil {
						logger.Err(err)
						return
					}
					for _, slowlog := range slowlogs {
						command := strings.ToLower(slowlog.Args[0])
						if CmdsToIgnore[command] {
							continue
						}
						logger.Info().
							Str("type", "SLOWLOG").
							Str("client-addr", slowlog.ClientAddr).
							Str("client-name", clientList[slowlog.ClientAddr]).
							Str("duration", slowlog.Duration.String()).
							Msg(fmt.Sprint(slowlog.Args))
					}
				}()
			}
		}
	}()

	// get client list
	wg.Add(1)
	go func() {
		defer wg.Done()
		tick := time.Tick(3000 * time.Millisecond)
		cmd := rdb.ClientList(ctx)

		for {
			select {
			case <-ctx.Done():
				logger.Info().Msg("stop getting client list")
				return
			case <-tick:
				rdb.Process(ctx, cmd)
				result, err := cmd.Result()
				if err != nil {
					logger.Err(err)
					continue
				}
				// parse result
				clients := strings.Split(result, "\n")
				for _, client := range clients {
					properties := strings.Split(client, " ")
					clientIP := ""
					clientUsername := ""
					for _, property := range properties {
						if len(clientIP) != 0 && len(clientUsername) != 0 {
							break
						}
						tmp := strings.Split(property, "=")
						if len(tmp) < 2 {
							continue
						}
						if tmp[0] == "addr" {
							clientIP = tmp[1]
						} else if tmp[0] == "user" {
							clientUsername = tmp[1]
						}
					}
					clientList[clientIP] = clientUsername
				}
			}
		}
	}()

	wg.Wait()
}
