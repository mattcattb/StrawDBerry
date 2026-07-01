package redis

import (
	"strconv"
	"strings"
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

func tryParseInt64(val string) (int64, bool) {
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, false
	}

	return n, true
}

func newStringObject(value string) *RedisObject {

	if n, ok := tryParseInt(value); ok {
		return &RedisObject{
			typ:       StringObject,
			encoding:  EncodingInt,
			ptr:       intStringPayload(n),
			expiresAt: noExpiration,
		}
	}

	return &RedisObject{
		typ:       StringObject,
		encoding:  EncodingRaw,
		ptr:       rawStringPayload(value),
		expiresAt: noExpiration,
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

type setOptions struct {
	nx  bool
	xx  bool
	get bool
	ex  *int64 // expire in s
	px  *int64 // expire in ms

	keepTtl bool
	// ex expire seconds, px expire ms
	xAt  *int64 // expire at s
	pXAt *int64 // expire at ms

}

func parseOptionalArgs(args []string) (so setOptions, err error) {
	argLen := len(args)
	for i := 2; i < argLen; i += 1 {
		switch strings.ToUpper(args[i]) {
		case "NX":
			if so.xx {
				// xx has already been set, cannot set both (mutually exclusive)
				return so, ErrWrongArgs
			}
			so.nx = true
		case "XX":

			if so.nx {
				return so, ErrWrongArgs
			}

			so.xx = true

		case "EX":
			if i+1 >= argLen {
				// we dont have an extra one here
				return so, ErrWrongArgs
			}

			secStr := args[i+1]
			secVal, ok := tryParseInt64(secStr)
			if !ok {
				return so, ErrInvalidEncoding
			}

			so.ex = &secVal
			i++

		case "PX":
			if i+1 >= argLen {
				// we dont have an extra one here
				return so, ErrWrongArgs
			}

			msStr := args[i+1]
			msVal, ok := tryParseInt64(msStr)
			if !ok {
				return so, ErrInvalidEncoding
			}

			so.px = &msVal
			i++
		case "KEEPTTL":
			so.keepTtl = true
		}
	}

	return so, nil
}

func Set(c *Client, args []string) CommandResult {
	key, val := args[0], args[1]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	expiresAt := noExpiration

	options, err := parseOptionalArgs(args)

	if err != nil {
		return Failed(wrongArgs("SET"))
	}

	if options.ex != nil {
		expiresAt = time.Now().UnixMilli() + (*options.ex * 1000)
	}
	if options.px != nil {
		expiresAt = time.Now().UnixMilli() + *options.px
	}

	newObj := newStringObject(val)
	newObj.expiresAt = expiresAt

	shouldSet := true

	if options.xx {
		// set if exists
		if _, exists := c.db.lookupKey(key); !exists {
			shouldSet = false
		}

	} else if options.nx {
		// set if doesnt exist

		if _, exists := c.db.lookupKey(key); exists {
			shouldSet = false
		}
	}

	if shouldSet {
		c.db.setKey(key, newObj)
		c.server.dirty += 1
		return Result(SimpleString("OK"))
	}
	// TODO get

	// GET: return old string stored at the key, or nil if the key did not exist. An error is returned and SET is aborted if the value stored at the key is not a string.

	return Result(Null())

}

func Get(c *Client, args []string) CommandResult {
	key := args[0]
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

func Incr(c *Client, args []string) CommandResult {

	key := args[0]

	finalVal, err := deltaStrValue(c, key, 1)

	if err != nil {
		return Failed(wrongTypeError())
	}
	c.server.dirty += 1
	return Result(Integer(finalVal))
}

func Decr(c *Client, args []string) CommandResult {

	key := args[0]
	finalVal, err := deltaStrValue(c, key, -1)

	if err != nil {
		return Failed(wrongTypeError())
	}
	c.server.dirty += 1
	return Result(Integer(finalVal))
}

func DecrBy(c *Client, args []string) CommandResult {
	key, decrByArg := args[0], args[1]
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

func IncrBy(c *Client, args []string) CommandResult {
	key, incrByArg := args[0], args[1]
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

func MGet(c *Client, args []string) CommandResult {
	respValues := make([]Value, 0)

	for _, key := range args {
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

func MSet(c *Client, args []string) CommandResult {
	// MSET key value [key value ...]
	if len(args)%2 == 1 {
		return Failed(wrongArgs("mset"))
	}

	kvPairs := make([][2]string, 0)

	for i := 0; i < len(args); i += 2 {
		kArg, valArg := args[i], args[i+1]
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

// Returns the length of the string value stored at key. An error is returned when key holds a non-string value.
func StrLen(c *Client, args []string) CommandResult {
	key := args[0]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(0))
	}

	vStr, n := stringObjectValue(obj)

	if n != nil {
		return Failed(wrongTypeError())
	}

	return Result(Integer(len(vStr)))
}

type lcsArgs struct {
	length       bool
	idx          bool
	minMatchLen  int  // When used with IDX, return matches at least min-match-len char long.
	withMatchLen bool // include the length of each match in the result
}

func Lcs(c *Client, args []string) CommandResult {
	// TODO
	//The LCS command implements the longest common subsequence algorithm. Note that this is different than the longest common string algorithm, since matching characters in the string does not need to be contiguous.
	// LCS key1 key2 [LEN] [IDX] [MINMATCHLEN min-match-len] [WITHMATCHLEN]
	return Failed(SimpleError("not yet implemented"))
}
