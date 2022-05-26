package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func main() {
	// init
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	defer rdb.Close()

	// test connection
	fmt.Printf("PING -> ")
	ping := rdb.Ping(ctx)
	result, err := ping.Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)

	fmt.Printf("GET hello -> ")
	result, err = rdb.Get(ctx, "hello").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)

	pubsub := rdb.Subscribe(ctx, "__key*__:*")
	defer pubsub.Close()

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			panic(err)
		}
		fmt.Println(msg.Channel)
	}
}
