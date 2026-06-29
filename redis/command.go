package redis

import (
	"errors"
	"strings"
)

type CommandResult struct {
	Reply Value

	Failed bool
}

type CommandContext = Client

func NewExec(db *RedisDb) *CommandContext {
	return &CommandContext{db: db}
}

func Result(reply Value) CommandResult {
	return CommandResult{Reply: reply}
}

func Failed(reply Value) CommandResult {
	return CommandResult{Reply: reply, Failed: true}
}

// HMMM lets try to get structured commands here
/*
func Execute(e *CommandContext, v Value) (res CommandResult) {

	command, args, ok := ParseCommand(v)

	if !ok {
		return Failed(syntaxError())
	}

	spec, exists := c.

	if !exists {
		return Failed(unknownCommand(command))
	}

	if err := parseSpec(spec, args); err != nil {
		return Failed(wrongArgs(command))
	}

	// for aof persistance
	// writes := spec.write

	return spec.handler(e, args)
}
*/

func ParseCommand(v Value) (string, []Value, bool) {
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

func wrongArgs(command string) Value {
	return Error("ERR wrong number of arguments for '" + command + "' command")
}

func syntaxError() Value {
	return Error("ERR syntax error")
}

func invalidInteger() Value {
	return Error("ERR value is not an integer or out of range")
}

func wrongTypeError() Value {
	return Error("WRONGTYPE Operation against a key holding the wrong kind of value")
}

func unknownCommand(command string) Value {
	return Error("ERR unknown command '" + command + "'")
}

func internalError() Value {
	return Error("ERR internal server error")
}

func mapRedisError(err error) Value {

	switch {
	case errors.Is(err, ErrWrongType):
		return wrongTypeError()

	case errors.Is(err, ErrInvalidEncoding):
		return wrongTypeError()
	}

	return internalError()

}

func parseBulkRespStringCommands(args []Value) ([]string, error) {
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
