package redis

import (
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

func ParseCommand(v Value) (string, []string, error) {
	values, ok := v.Array()
	if !ok || len(values) == 0 {
		return "", nil, ErrWrongArgs
	}

	commandName, ok := values[0].BulkString()
	if !ok {
		return "", nil, ErrWrongArgs
	}

	argVals, err := parseBulkStrCommand(values[1:])

	if err != nil {
		return "", nil, ErrWrongArgs
	}

	return strings.ToUpper(commandName), argVals, nil
}

func parseBulkStrCommand(args []Value) ([]string, error) {
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
