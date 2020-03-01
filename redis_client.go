package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"os"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"), // no password set
		DB:       0,                           // use default DB
	})

	_, err := redisClient.Ping().Result()

	if err != nil {
		panic(err)
	}

	fmt.Println("connected to redis")
}
