package utils

import (
	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis"
	"sync"
)

var (
	node     *snowflake.Node
	nodeOnce sync.Once
	rdb      *redis.Client
)

func Client() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	_, err := rdb.Ping().Result()
	if err != nil {
		panic(err)
	}
	return rdb
}
