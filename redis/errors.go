package redis

import "errors"

/*
 Type	Common Causes	When to Handle	Examples
Connection errors	Network issues, server down, auth failure,
timeouts, pool exhaustion	Almost always	ConnectionError, TimeoutError,
AuthenticationError
Command errors	Typo in command, wrong arguments, invalid types,
 unsupported command	Rarely (usually indicates a bug)
 ResponseError, WRONGTYPE, ERR unknown command
Data errors	Serialization failures, corrupted data,
type mismatches	Sometimes (depends on data source)
JSONDecodeError, SerializationError, WRONGTYPE
Resource errors	Memory limit, pool exhausted,
too many connections, key eviction	Sometimes (some are temporary)
OOM, pool timeout, LOADING
*/

// Connection Errors
var ErrConnection = errors.New("ConnectionError")
var ErrTimeout = errors.New("TimeoutError")
var ErrAuthentication = errors.New("AuthenticationError")

// Command Errors
var ErrInvalidEncoding = errors.New(("wrong encoding"))
var ErrWrongArgs = errors.New("wrong number of arguments")
var ErrUnknownCommand = errors.New("ERR unknown command")
var ErrWrongType = errors.New("WRONGTYPE") // wrong object type

var ErrInvalidState = errors.New("INVALID CLIENT STATE")

// Data Serialization Errors
var ErrJsonDecode = errors.New("JSONDecodeError")
var ErrSerialization = errors.New("SerializationError")

// Resource Liits
var ErrMemoryLimit = errors.New("ERRMemoryLimit")

var ErrInternal = errors.New("InternalError")

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

	switch err {
	case ErrWrongArgs:
		return wrongTypeError()
	}

	switch {
	case errors.Is(err, ErrWrongType):
		return wrongTypeError()

	case errors.Is(err, ErrInvalidEncoding):
		return wrongTypeError()
	}

	return internalError()
}
