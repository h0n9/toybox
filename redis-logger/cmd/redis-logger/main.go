package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/h0n9/toybox/redis-logger/util"
	"github.com/rs/zerolog/log"
)

const (
	DefaultServiceName   = "redis-logger"
	DefaultRedisAddr     = "localhost:6379"
	DefaultRedisPassword = ""
)

func main() {
	// get envs
	serviceName := util.GetEnv("SERVICE_NAME", DefaultServiceName)
	addr := util.GetEnv("REDIS_ADDR", DefaultRedisAddr)
	password := util.GetEnv("REDIS_PASSWORD", DefaultRedisPassword)

	// init
	ctx := context.Background()
	wg := sync.WaitGroup{}
	logger := log.With().
		Str("service", serviceName).
		Str("redis-addr", addr).
		Logger()
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	defer rdb.Close()

	// test connection
	ping := rdb.Ping(ctx)
	result, err := ping.Result()
	if err != nil {
		logger.Panic().Msg(err.Error())
	}
	logger.Info().Msg(result)

	// subscribe keyevent
	wg.Add(1)
	go func() {
		defer wg.Done()

		pubsub := rdb.PSubscribe(ctx, "__keyevent*__:*")
		defer pubsub.Close()

		for {
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				logger.Panic().Msg(err.Error())
			}
			logger.Info().Msg(msg.Channel)
		}
	}()

	// get slowlogs
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			slowlogs, err := rdb.SlowLogGet(ctx, 1).Result()
			if err != nil {
				logger.Err(err)
				continue
			}
			if len(slowlogs) < 1 {
				continue
			}
			logger.Info().Msg(fmt.Sprint(slowlogs[0].Args))
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}
