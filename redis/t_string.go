package redis

import (
	"strconv"
	"time"
)

// register t_string commands + context to the handler

func tryParseInt(val string) (int, bool) {
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return n, true

}

func registerTstringCommands(sh *SpecHandler) {

	specMap := map[string]CommandSpec{
		"GET": CommandSpec{
			arity:   1,
			flags:   CmdRead,
			handler: Get,
		},
		"SET": CommandSpec{
			arity:   -2,
			flags:   CmdWrite,
			handler: Set,
		},
		"MGET": CommandSpec{
			arity:   -1,
			flags:   CmdRead,
			handler: MGet,
		},

		"MSET": CommandSpec{
			arity:   -2,
			flags:   CmdWrite,
			handler: MSet,
		},
		"INCR": {
			arity:   1,
			handler: Incr,
			flags:   CmdWrite,
		},
		"DECR": {
			arity:   1,
			handler: Decr,
			flags:   CmdWrite,
		},

		"INCRBY": {
			arity:   2,
			handler: IncrBy,
			flags:   CmdWrite,
		},

		"DECRBY": {
			arity:   2,
			handler: DecrBy,
			flags:   CmdWrite,
		},
	}

	for k, v := range specMap {
		v.group = StringGroup
		v.name = k
		specMap[k] = v
	}

	sh.registerCommandSpecs(specMap)
}

func newStringObject(value string) *RedisObject {

	if n, ok := tryParseInt(value); ok {
		return &RedisObject{
			typ:      StringObject,
			encoding: EncodingInt,
			ptr:      intStringPayload(n),
		}
	}

	return &RedisObject{
		typ:      StringObject,
		encoding: EncodingRaw,
		ptr:      rawStringPayload(value),
	}
}

func stringObjectValue(obj *RedisObject) (string, error) {
	if obj.typ != StringObject {
		return "", ErrWrongType
	}

	switch obj.encoding {
	case EncodingRaw:
		value, ok := obj.ptr.(rawStringPayload)
		if !ok {
			return "", ErrInvalidEncoding
		}
		return string(value), nil

	case EncodingInt:
		value, ok := obj.ptr.(intStringPayload)
		if !ok {
			return "", ErrInvalidEncoding
		}
		return strconv.Itoa(int(value)), nil
	}

	return "", ErrInvalidEncoding
}

func setStringObjectValue(obj *RedisObject, value string) {
	if n, ok := tryParseInt(value); ok {
		obj.encoding = EncodingInt
		obj.ptr = intStringPayload(n)
		return
	}

	obj.encoding = EncodingRaw
	obj.ptr = rawStringPayload(value)
}

func Set(c *Client, args []Value) CommandResult {
	key, ok := args[0].BulkString()
	if !ok {
		return Failed(syntaxError())
	}
	val, ok := args[1].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	// NX Only set the key if it does not already exist.

	// XX Only set the key if it already exists.

	setBehavior := 0 // -1 doesnt already exist, 1 set if already exists
	expiresAtMs := -1

	for i := 2; i < len(args); i += 1 {
		arg := args[i]
		argVal, ok := arg.BulkString()

		if !ok {
			return Failed(wrongArgs("set"))
		}
		switch argVal {
		case "EX", "ex":
			// expire at s
			if len(args) < i+1 {
				return Failed(wrongArgs("set"))
			}
			expireSArg, ok := args[i+1].BulkString()

			if !ok {
				return Failed(syntaxError())
			}

			expireAtS, ok := tryParseInt(expireSArg)
			if !ok || (expireAtS < 1) {
				return Failed(invalidInteger())
			}
			expiresAtMs = expireAtS * 1000
			i += 1
			continue
		case "PX", "px":
			// expire at ms
			if len(args) < i+1 {
				return Failed(wrongArgs("set"))
			}
			expireMSArg, ok := args[i+1].BulkString()

			if !ok {
				return Failed(syntaxError())
			}

			expireAtMS, ok := tryParseInt(expireMSArg)
			if !ok || (expireAtMS < 1) {
				return Failed(invalidInteger())
			}
			expiresAtMs = expireAtMS
			i += 1
			continue
		case "NX", "nx":
			// only set if doesnt exist
			setBehavior = -1
			continue
		case "XX", "xx":
			// only set if exists
			setBehavior = 1
			continue

		default:
			return Failed(syntaxError())
		}
	}

	newObj := newStringObject(val)

	if expiresAtMs != -1 {
		newObj.expiresAt = time.Now().Add(time.Millisecond * time.Duration(expiresAtMs))
	}

	shouldSet := false

	if setBehavior != 0 {
		_, exists := c.db.lookupKey(key)
		if setBehavior == -1 && !exists {
			// set if doesnt already exist

			shouldSet = true
		}
		if setBehavior == 1 && exists {
			// set if already exists
			shouldSet = true
		}
	} else {
		shouldSet = true
	}

	if shouldSet {
		c.db.setKey(key, newObj)
		c.server.dirty += 1
		return Result(SimpleString("OK"))
	}

	return Result(Null())

}

func Get(c *Client, args []Value) CommandResult {

	if len(args) != 1 {
		return Failed(wrongArgs("get"))
	}

	keyArg := args[0]

	key, ok := keyArg.BulkString()

	if !ok {
		return Failed(wrongArgs("get"))
	}

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Null())
	}

	val, err := stringObjectValue(obj)

	if err != nil {
		return Failed(wrongTypeError())
	}

	return Result(BulkString(val))
}

func Incr(c *Client, args []Value) CommandResult {

	if len(args) != 1 {
		return Failed(wrongArgs("incr"))
	}

	key, ok := args[0].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	finalVal, err := deltaStrValue(c, key, 1)

	if err != nil {
		return Failed(wrongTypeError())
	}
	c.server.dirty += 1
	return Result(Integer(finalVal))
}

func Decr(c *Client, args []Value) CommandResult {
	if len(args) != 1 {
		return Failed(wrongArgs("decr"))
	}

	key, ok := args[0].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	finalVal, err := deltaStrValue(c, key, -1)

	if err != nil {
		return Failed(wrongTypeError())
	}
	c.server.dirty += 1
	return Result(Integer(finalVal))
}

func DecrBy(c *Client, args []Value) CommandResult {

	if len(args) != 2 {
		return Failed(wrongArgs("decrby"))
	}

	key, ok := args[0].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	decrByArg, ok := args[1].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	decrBy, err := strconv.Atoi(decrByArg)

	if err != nil {
		return Failed(invalidInteger())
	}

	n, err := deltaStrValue(c, key, decrBy*-1)

	if err != nil {
		return Failed(wrongTypeError())
	}

	c.server.dirty += 1
	return Result(Integer(n))
}

func IncrBy(c *Client, args []Value) CommandResult {
	if len(args) != 2 {
		return Failed(wrongArgs("incrby"))
	}

	key, ok := args[0].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	incrByArg, ok := args[1].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

	incrBy, err := strconv.Atoi(incrByArg)

	if err != nil {
		return Failed(invalidInteger())
	}

	n, err := deltaStrValue(c, key, incrBy)

	if err != nil {
		return Failed(wrongTypeError())
	}

	c.server.dirty += 1
	return Result(Integer(n))
}

func deltaStrValue(c *Client, key string, delta int) (int, error) {
	// add/subtract value from here
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	curObj, _ := c.db.lookupKey(key)

	value := delta

	if curObj != nil {
		// cur obj exists
		strVal, ok := stringObjectValue(curObj)

		if ok != nil || curObj.typ != StringObject {
			return 0, ErrWrongType
		}

		n, err := strconv.Atoi(strVal)
		if err != nil {
			return 0, ErrWrongType
		}

		value += n

		setStringObjectValue(curObj, strconv.Itoa(value))
	} else {
		c.db.setKey(key, newStringObject(strconv.Itoa(value)))
	}

	return value, nil
}

func MGet(c *Client, args []Value) CommandResult {
	if len(args) < 1 {
		return Failed(wrongArgs("mget"))
	}

	mKeys := make([]string, len(args))
	respValues := make([]Value, 0)

	mKeys, err := parseBulkRespStringCommands(args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	for _, key := range mKeys {
		obj, exists := c.db.lookupKey(key)
		if !exists {
			respValues = append(respValues, Null())
			continue
		}

		strVal, err := stringObjectValue(obj)
		if err != nil {
			respValues = append(respValues, Null())
		} else {
			respValues = append(respValues, BulkString(strVal))
		}
	}

	return Result(Array(respValues))

}

func MSet(c *Client, args []Value) CommandResult {
	// MSET key value [key value ...]

	if len(args) < 2 {
		return Failed(wrongArgs("mset"))
	}

	if len(args)%2 == 1 {
		return Failed(wrongArgs("mset"))
	}

	kvPairs := make([][2]string, 0)

	for i := 0; i < len(args); i += 2 {
		kArg, ok := args[i].BulkString()
		if !ok {
			return Failed(syntaxError())
		}
		valArg, ok := args[i+1].BulkString()
		if !ok {
			return Failed(syntaxError())
		}
		kvPairs = append(kvPairs, [2]string{kArg, valArg})
	}

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	for _, pair := range kvPairs {
		key, val := pair[0], pair[1]
		c.db.setKey(key, newStringObject(val))
	}

	c.server.dirty += 1
	return Result(SimpleString("ok"))
}
