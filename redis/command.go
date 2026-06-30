package redis

import (
	"errors"
	"strings"
)

type CommandResult struct {
	Reply  Value
	Failed bool
}

func Result(reply Value) CommandResult {
	return CommandResult{Reply: reply}
}

func Failed(reply Value) CommandResult {
	return CommandResult{Reply: reply, Failed: true}
}

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

func invalidStateError() Value {
	return Error("INVALID STATE for current client mode")
}

func unknownCommand(command string) Value {
	return Error("ERR unknown command '" + command + "'")
}

func internalError() Value {
	return Error("ERR internal server error")
}

func mapRedisErrorToResp(err error) Value {

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
