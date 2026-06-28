package server

import (
	"go-redis/persistance"
	"go-redis/pubsub"
	"go-redis/redis"
)

type Server struct {
	db     *redis.RedisDb
	exec   *redis.CommandExecutor
	pubsub *pubsub.PubSub
	aof    *persistance.Aof
}
