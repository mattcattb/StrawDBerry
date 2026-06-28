package redis

import (
	"errors"
	"go-redis/resp"
	"strings"
)

type appendLog interface {
	Append(resp.Value) error
}

type CommandExecutor struct {
	db  *RedisDb
	aof appendLog
}

// HMMM lets try to get structured commands here

func (e *CommandExecutor) Execute(v resp.Value) resp.Value {

	command, args, ok := parseCommand(v)

	if !ok {
		return syntaxError()
	}

	spec, exists := handler[command]

	if !exists {
		return unknownCommand(command)
	}

	if err := parseSpec(spec, args); err != nil {
		return wrongArgs(command)
	}

	// for aof persistance
	// writes := spec.write

	return spec.handler(e, args)
}

func parseCommand(v resp.Value) (string, []resp.Value, bool) {
	values, ok := v.Array()
	if !ok || len(values) == 0 {
		return "", nil, false
	}

	commandName, ok := values[0].BulkString()
	if !ok {
		return "", nil, false
	}

	return strings.ToUpper(commandName), values[1:], true
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

func mapRedisError(err error) resp.Value {

	switch {
	case errors.Is(err, ErrWrongType):
		return wrongTypeError()

	case errors.Is(err, ErrInvalidEncoding):
		return wrongTypeError()
	}

	return internalError()

}
