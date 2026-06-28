package redis

import (
	"errors"
	"go-redis/resp"
)

var errWrongArgs = errors.New("wrong number of arguments")

func parseSpec(spec commandSpec, args []resp.Value) error {
	minArgs, maxArgs := spec.minArgs, spec.maxArgs

	if minArgs > 0 && len(args) < minArgs {
		return errWrongArgs
	}

	if maxArgs >= 0 && len(args) > maxArgs {
		return errWrongArgs
	}

	return nil
}

func parseBulkRespStringCommands(args []resp.Value) ([]string, error) {
	// parse all args into a string array

	stringArgs := make([]string, len(args))

	for i := 0; i < len(args); i += 1 {
		strVal, ok := args[i].BulkString()
		if !ok {
			return stringArgs, ErrWrongType
		}
		stringArgs[i] = strVal
	}

	return stringArgs, nil

}

type commandHandler func(*CommandExecutor, []resp.Value) resp.Value

type commandSpec struct {
	minArgs int
	maxArgs int
	write   bool
	handler commandHandler
}

var handler = map[string]commandSpec{
	"EXISTS": {
		handler: (*CommandExecutor).Exists,
		minArgs: 1,
		maxArgs: -1,
		write:   false,
	},
	"SET": {
		minArgs: 2,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).Set,
	},
	"GET": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).Get,
	},
	"ECHO": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).Echo,
	},
	"TYPE": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).Type,
	},
	"TTL": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).Ttl,
	},
	"PING": {
		minArgs: 0,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).Ping,
	},
	"DEL": {
		minArgs: 1,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).Del,
	},
	"MGET": {
		minArgs: 1,
		maxArgs: -1,
		write:   false,
		handler: (*CommandExecutor).MGet,
	},
	"MSET": {
		minArgs: 2,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).MSet,
	},
	"INCR": {
		minArgs: 1,
		maxArgs: 1,
		write:   true,
		handler: (*CommandExecutor).Incr,
	},
	"DECR": {
		minArgs: 1,
		maxArgs: 1,
		write:   true,
		handler: (*CommandExecutor).Decr,
	},
	"INCRBY": {
		minArgs: 2,
		maxArgs: 2,
		write:   true,
		handler: (*CommandExecutor).IncrBy,
	},
	"DECRBY": {
		minArgs: 2,
		maxArgs: 2,
		write:   true,
		handler: (*CommandExecutor).DecrBy,
	},
	"HGET": {
		minArgs: 2,
		maxArgs: 2,
		write:   false,
		handler: (*CommandExecutor).HGet,
	},
	"HSET": {
		minArgs: 3,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).HSet,
	},
	"HDEL": {
		minArgs: 2,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).HDel,
	},
	"HGETALL": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).HGetAll,
	},
	"HEXISTS": {
		minArgs: 2,
		maxArgs: 2,
		write:   false,
		handler: (*CommandExecutor).HExists,
	},
	"SADD": {
		minArgs: 2,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).SAdd,
	},
	"SCARD": {
		minArgs: 1,
		maxArgs: 1,
		write:   false,
		handler: (*CommandExecutor).SCard,
	},
	"SREM": {
		minArgs: 2,
		maxArgs: -1,
		write:   true,
		handler: (*CommandExecutor).SRem,
	},
	"SISMEMBER": {
		minArgs: 2,
		maxArgs: 2,
		write:   false,
		handler: (*CommandExecutor).SIsMem,
	},
}
