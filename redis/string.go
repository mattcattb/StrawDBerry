package redis

import (
	"go-redis/resp"
	"strconv"
	"time"
)

func tryParseInt(val string) (int, bool) {
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, false
	}

	return n, true

}

func newStringObject(value string) *RedisObject {

	if n, ok := tryParseInt(value); ok {
		return &RedisObject{
			typ:      StringObject,
			encoding: EncodingInt,
			ptr:      n,
		}
	}

	return &RedisObject{
		typ:      StringObject,
		encoding: EncodingRaw,
		ptr:      value,
	}
}

func stringObjectValue(obj *RedisObject) (string, error) {
	if obj.typ != StringObject {
		return "", ErrWrongType
	}

	switch obj.encoding {
	case EncodingRaw:
		value, ok := obj.ptr.(string)
		if !ok {
			return "", ErrInvalidEncoding
		}
		return value, nil

	case EncodingInt:
		value, ok := obj.ptr.(int64)
		if !ok {
			return "", ErrInvalidEncoding
		}
		return strconv.FormatInt(value, 10), nil
	}

	return "", ErrInvalidEncoding
}

func setStringObjectValue(obj *RedisObject, value string) {
	if n, ok := tryParseInt(value); ok {
		obj.encoding = EncodingInt
		obj.ptr = n
		return
	}

	obj.encoding = EncodingRaw
	obj.ptr = value
}

func (ce *CommandExecutor) Set(args []resp.Value) resp.Value {
	if len(args) < 2 {
		return wrongArgs("set")
	}

	key, ok := args[0].BulkString()
	if !ok {
		return syntaxError()
	}
	val, ok := args[1].BulkString()

	if !ok {
		return syntaxError()
	}

	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	// NX Only set the key if it does not already exist.

	// XX Only set the key if it already exists.

	setBehavior := 0 // -1 doesnt already exist, 1 set if already exists
	expiresAtMs := -1

	for i := 2; i < len(args); i += 1 {
		arg := args[i]
		argVal, ok := arg.BulkString()

		if !ok {
			return wrongArgs("set")
		}
		switch argVal {
		case "EX", "ex":
			// expire at s
			if len(args) < i+1 {
				return wrongArgs("set")
			}
			expireSArg, ok := args[i+1].BulkString()

			if !ok {
				return syntaxError()
			}

			expireAtS, ok := tryParseInt(expireSArg)
			if !ok || (expireAtS < 1) {
				return invalidInteger()
			}
			expiresAtMs = expireAtS * 1000
			i += 1
			continue
		case "PX", "px":
			// expire at ms
			if len(args) < i+1 {
				return wrongArgs("set")
			}
			expireMSArg, ok := args[i+1].BulkString()

			if !ok {
				return syntaxError()
			}

			expireAtMS, ok := tryParseInt(expireMSArg)
			if !ok || (expireAtMS < 1) {
				return invalidInteger()
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
			return syntaxError()
		}
	}

	newObj := newStringObject(val)

	if expiresAtMs != -1 {
		newObj.expiresAt = time.Now().Add(time.Millisecond * time.Duration(expiresAtMs))
	}

	shouldSet := false

	if setBehavior != 0 {
		_, exists := ce.db.lookupKey(key)
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
		ce.db.setKey(key, newObj)
		return resp.SimpleString("OK")
	}

	return resp.Null()

}

func (ce *CommandExecutor) Get(args []resp.Value) resp.Value {

	if len(args) != 1 {
		return wrongArgs("get")
	}

	keyArg := args[0]

	key, ok := keyArg.BulkString()

	if !ok {
		return wrongArgs("get")
	}

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Null()
	}

	val, err := stringObjectValue(obj)

	if err != nil {
		return wrongTypeError()
	}

	return resp.BulkString(val)
}

func (ce *CommandExecutor) Incr(args []resp.Value) resp.Value {

	if len(args) != 1 {
		return wrongArgs("incr")
	}

	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	finalVal, err := ce.deltaStrValue(key, 1)

	if err != nil {
		return wrongTypeError()
	}
	return resp.Integer(finalVal)
}

func (ce *CommandExecutor) Decr(args []resp.Value) resp.Value {
	if len(args) != 1 {
		return wrongArgs("decr")
	}

	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	finalVal, err := ce.deltaStrValue(key, -1)

	if err != nil {
		return wrongTypeError()
	}
	return resp.Integer(finalVal)
}

func (ce *CommandExecutor) DecrBy(args []resp.Value) resp.Value {

	if len(args) != 2 {
		return wrongArgs("decrby")
	}

	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	decrByArg, ok := args[1].BulkString()

	if !ok {
		return syntaxError()
	}

	decrBy, err := strconv.Atoi(decrByArg)

	if err != nil {
		return invalidInteger()
	}

	n, err := ce.deltaStrValue(key, decrBy*-1)

	if err != nil {
		return wrongTypeError()
	}

	return resp.Integer(n)
}

func (ce *CommandExecutor) IncrBy(args []resp.Value) resp.Value {
	if len(args) != 2 {
		return wrongArgs("incrby")
	}

	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	incrByArg, ok := args[1].BulkString()

	if !ok {
		return syntaxError()
	}

	incrBy, err := strconv.Atoi(incrByArg)

	if err != nil {
		return invalidInteger()
	}

	n, err := ce.deltaStrValue(key, incrBy)

	if err != nil {
		return wrongTypeError()
	}

	return resp.Integer(n)
}

func (ce *CommandExecutor) deltaStrValue(key string, delta int) (int, error) {
	// add/subtract value from here
	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	curObj, _ := ce.db.lookupKey(key)

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
		ce.db.setKey(key, newStringObject(strconv.Itoa(value)))
	}

	return value, nil
}

func (ce *CommandExecutor) MGet(args []resp.Value) resp.Value {
	if len(args) < 1 {
		return wrongArgs("mget")
	}

	mKeys := make([]string, len(args))
	respValues := make([]resp.Value, 0)

	mKeys, err := parseBulkRespStringCommands(args)
	if err != nil {
		return wrongTypeError()
	}

	for _, key := range mKeys {
		obj, exists := ce.db.lookupKey(key)
		if !exists {
			respValues = append(respValues, resp.Null())
			continue
		}

		strVal, err := stringObjectValue(obj)
		if err != nil {
			respValues = append(respValues, resp.Null())
		} else {
			respValues = append(respValues, resp.BulkString(strVal))
		}
	}

	return resp.Array(respValues)

}

func (ce *CommandExecutor) MSet(args []resp.Value) resp.Value {
	// MSET key value [key value ...]

	if len(args) < 2 {
		return wrongArgs("mset")
	}

	if len(args)%2 == 1 {
		return wrongArgs("mset")
	}

	kvPairs := make([][2]string, 0)

	for i := 0; i < len(args); i += 2 {
		kArg, ok := args[i].BulkString()
		if !ok {
			return syntaxError()
		}
		valArg, ok := args[i+1].BulkString()
		if !ok {
			return syntaxError()
		}
		kvPairs = append(kvPairs, [2]string{kArg, valArg})
	}

	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	for _, pair := range kvPairs {
		key, val := pair[0], pair[1]
		ce.db.setKey(key, newStringObject(val))
	}

	return resp.SimpleString("ok")
}
