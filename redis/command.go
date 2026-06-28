package redis

import (
	"go-redis/persistance"
	"go-redis/resp"
	"strings"
)

type CommandType string

const (
	CommandPing   CommandType = "PING"
	CommandExists CommandType = "EXISTS"
	CommandExpire CommandType = "EXPIRE"
	CommandTtl    CommandType = "TTL"
	CommandInfo   CommandType = "INFO"
	CommandSet    CommandType = "SET"
	CommandDel    CommandType = "DEL"

	CommandGet  CommandType = "GET"
	CommandHSet CommandType = "HSET"
	CommandIncr CommandType = "INCR"
	CommandMGet CommandType = "MGET"
	CommandHGet CommandType = "HGET"
	CommandHDel CommandType = "HDEL"
)

type CommandExecutor struct {
	db  *RedisDb
	aof *persistance.Aof
}

// HMMM lets try to get structured commands here

func (e *CommandExecutor) Execute(v resp.Value) resp.Value {

	command, args, ok := parseCommand(v)

	if !ok {
		return syntaxError()
	}

	switch CommandType(command) {

	case CommandPing:
		return e.Ping(args)

	case CommandHSet:
		return e.HSet(args)
	case CommandSet:

		return e.Set(args)

	case CommandGet:
		return e.Get(args)

	case CommandMGet:
		return e.MGet(args)

	case CommandHGet:
		return e.HGet(args)

	case CommandDel:
		return e.Del(args)

	}

	return unknownCommand(string(command))
}

func parseCommand(v resp.Value) (CommandType, []resp.Value, bool) {
	values, ok := v.Array()
	if !ok || len(values) == 0 {
		return "", nil, false
	}

	commandName, ok := values[0].BulkString()
	if !ok {
		return "", nil, false
	}

	return CommandType(strings.ToUpper(commandName)), values[1:], true
}

func wrongArgs(command string) resp.Value {
	return resp.Error("ERR wrong number of arguments for '" + command + "' command")
}

func syntaxError() resp.Value {
	return resp.Error("ERR syntax error")
}

func invalidInteger() resp.Value {
	return resp.Error("ERR value is not an integer or out of range")
}

func wrongTypeError() resp.Value {
	return resp.Error("WRONGTYPE Operation against a key holding the wrong kind of value")
}

func unknownCommand(command string) resp.Value {
	return resp.Error("ERR unknown command '" + command + "'")
}

func internalError() resp.Value {
	return resp.Error("ERR internal server error")
}

func parseBulkRespStringCommands(args []resp.Value) []string {
	// parse all args into a string array
}
