package redis

import (
	"go-redis/resp"
	"time"
)

func (ce *CommandExecutor) Exists(args []resp.Value) resp.Value {
	existsKeys := make([]string, len(args))

	for i := 0; i < len(args); i += 1 {
		delKey, ok := args[i].BulkString()
		if !ok {
			return syntaxError()
		}

		existsKeys[i] = delKey
	}

	existCount := 0

	for _, key := range existsKeys {
		_, exists := ce.db.lookupKey(key)
		if exists {
			existCount += 1
		}
	}

	return resp.Integer(existCount)

}

func (ce *CommandExecutor) Expire(args []resp.Value) {

}

func (ce *CommandExecutor) Ttl(args []resp.Value) resp.Value {

	/*
		Integer reply: TTL in seconds.
		Integer reply: -1 if the key exists but has no associated expiration.
		Integer reply: -2 if the key does not exist.
	*/

	key, ok := args[0].BulkString()
	if !ok {
		return syntaxError()
	}

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Integer(-2)
	}

	if obj.expiresAt.IsZero() {
		return resp.Integer(-1)
	}

	seconds := int(time.Until(obj.expiresAt).Seconds())
	if seconds < 0 {
		delete(ce.db.dict, key)
		return resp.Integer(-2)
	}

	return resp.Integer(seconds)
}

func (ce *CommandExecutor) Del(args []resp.Value) resp.Value {
	delKeys := make([]string, len(args))

	for i := 0; i < len(args); i += 1 {
		delK, ok := args[i].BulkString()

		if !ok {
			return syntaxError()
		}

		delKeys[i] = delK
	}

	deletedCount := 0

	for _, dk := range delKeys {
		existed := ce.db.deleteKey(dk)
		if existed {
			deletedCount += 1
		}
	}

	return resp.Integer(deletedCount)

}

func (ce *CommandExecutor) Ping(args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.SimpleString("pong")
	}

	argMsg, _ := args[0].BulkString()

	return resp.BulkString(argMsg)
}

func (ce *CommandExecutor) Echo(args []resp.Value) resp.Value

// string, list, set, zset, hash, stream, and vectorset.
func (ce *CommandExecutor) Type(args []resp.Value) resp.Value {
	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.SimpleString("none")
	}

	switch ObjectType(obj.typ) {
	case StringObject:
		return resp.SimpleString("string")

	case ListObject:
		return resp.SimpleString("list")

	case SetObject:
		return resp.SimpleString("set")
	case ZSetObject:
		return resp.SimpleString("zset")

	case HashObject:
		return resp.SimpleString("hash")
	}

	return resp.Error("Unknown type")
}
