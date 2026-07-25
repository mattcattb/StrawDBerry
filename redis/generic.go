package redis

import (
	"strconv"
	"strings"
)

func Keys(c *Client, args []string) CommandResult {
	keys := make([]Value, 0)
	for _, key := range c.db.keysSnapshot() {
		if globMatch(args[0], key) {
			keys = append(keys, BulkString(key))
		}
	}

	return Result(Array(keys))
}

func Scan(c *Client, args []string) CommandResult {
	cursor, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return Failed(invalidInteger())
	}

	pattern := "*"
	count := 10
	for i := 1; i < len(args); {
		switch strings.ToUpper(args[i]) {
		case "MATCH":
			if i+1 >= len(args) {
				return Failed(syntaxError())
			}
			pattern = args[i+1]
			i += 2
		case "COUNT":
			if i+1 >= len(args) {
				return Failed(syntaxError())
			}

			parsedCount, err := strconv.Atoi(args[i+1])
			if err != nil || parsedCount <= 0 {
				return Failed(invalidInteger())
			}
			count = parsedCount
			i += 2
		default:
			return Failed(syntaxError())
		}
	}

	matched := make([]string, 0)
	for _, key := range c.db.keysSnapshot() {
		if globMatch(pattern, key) {
			matched = append(matched, key)
		}
	}

	if cursor >= uint64(len(matched)) {
		return Result(Array([]Value{BulkString("0"), Array([]Value{})}))
	}

	end := cursor + uint64(count)
	if end > uint64(len(matched)) {
		end = uint64(len(matched))
	}

	keys := make([]Value, 0, end-cursor)
	for _, key := range matched[cursor:end] {
		keys = append(keys, BulkString(key))
	}

	nextCursor := "0"
	if end < uint64(len(matched)) {
		nextCursor = strconv.FormatUint(end, 10)
	}

	return Result(Array([]Value{BulkString(nextCursor), Array(keys)}))
}

func globMatch(pattern, key string) bool {
	patternRunes := []rune(pattern)
	keyRunes := []rune(key)

	var match func(int, int) bool
	match = func(patternIndex, keyIndex int) bool {
		if patternIndex == len(patternRunes) {
			return keyIndex == len(keyRunes)
		}

		switch patternRunes[patternIndex] {
		case '*':
			for nextKeyIndex := keyIndex; nextKeyIndex <= len(keyRunes); nextKeyIndex++ {
				if match(patternIndex+1, nextKeyIndex) {
					return true
				}
			}
			return false
		case '?':
			return keyIndex < len(keyRunes) && match(patternIndex+1, keyIndex+1)
		case '\\':
			if patternIndex+1 < len(patternRunes) {
				return keyIndex < len(keyRunes) && patternRunes[patternIndex+1] == keyRunes[keyIndex] && match(patternIndex+2, keyIndex+1)
			}
		}

		return keyIndex < len(keyRunes) && patternRunes[patternIndex] == keyRunes[keyIndex] && match(patternIndex+1, keyIndex+1)
	}

	return match(0, 0)
}

// OBJECT ENCODING

func Copy(c *Client, args []string) CommandResult {

	return Failed(Error("not implemented yet"))

	/*
		srcK, destK := args[0], args[1]
		replace := false

		// COPY source destination [REPLACE]
		res := 0 // 0 not copied, 1 if copied

		if len(args) > 2 && args[2] == "REPLACE" {
			replace = true
		}

		c.db.mu.Lock()
		defer c.db.mu.Unlock()

		src, e := c.db.lookupKey(srcK)

		if e {
			// e exists so lets get the resp
			dest, de := c.db.lookupKey(destK)

			// replace IF replace true OR dest empty

			if !de || replace {
				newVal := RedisObject{}
				copy(src, &newVal)
				c.db.setKey(destK, &newVal)
			}
		}

		return Result(Integer(res)) */
}

func Exists(c *Client, args []string) CommandResult {
	existCount := 0

	for _, key := range args {
		_, exists := c.db.lookupKey(key)
		if exists {
			existCount += 1
		}
	}

	return Result(Integer(existCount))

}

func Expire(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func Ttl(c *Client, args []string) CommandResult {

	/*
		Integer reply: TTL in seconds.
		Integer reply: -1 if the key exists but has no associated expiration.
		Integer reply: -2 if the key does not exist.
	*/

	key := args[0]

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(-2))
	}

	seconds := obj.ttlForObject()
	if seconds == -2 {
		delete(c.db.dict, key)
		c.server.dirty += 1
		return Result(Integer(-2))
	}

	return Result(Integer(int(seconds)))
}

func Persist(c *Client, args []string) CommandResult {
	// Remove the existing timeout on key, turning the key from volatile (a key with an expire set) to persistent (a key that will never expire as no timeout is associated).

	key := args[0]

	obj, e := c.db.lookupKey(key)

	rVal := 0

	// return int 0 if doenst exist or no timeout
	// 1 if persisted

	if e && obj != nil {
		obj.setExprMs(-1)
		rVal = 1
	}

	return Result(Integer(rVal))
}

func Del(c *Client, args []string) CommandResult {
	deletedCount := 0

	for _, dk := range args {
		existed := c.db.deleteKey(dk)
		if existed {
			deletedCount += 1
		}
	}

	if deletedCount > 0 {
		c.server.dirty += 1
	}

	return Result(Integer(deletedCount))

}

// string, list, set, zset, hash, stream, and vectorset.
func Type(c *Client, args []string) CommandResult {
	key := args[0]

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(SimpleString("none"))
	}
	str := obj.typ.Str()
	return Result(SimpleString(str))
}

// RDB based function
func Dump(c *Client, args []string) CommandResult {

	// value stored at key in a Redis-specific format and return it to the user. The returned value can be synthesized back into a Redis key using the RESTORE command
	/* It contains a 64-bit checksum that is used to make sure errors will be detected. The RESTORE command makes sure to check the checksum before synthesizing a key using the serialized value.
	Values are encoded in the same format used by RDB.
	An RDB version is encoded inside the serialized value, so that different Redis versions with incompatible RDB formats will refuse to process the serialized value.
	*/
	return Failed(Error("not implemented yet"))

}

func Restore(c *Client, args []string) CommandResult {
	/*
	   RESTORE key ttl serialized-value [REPLACE] [ABSTTL]

	   	[IDLETIME seconds] [FREQ frequency]
	*/
	return Failed(Error("not implemented yet"))
}

func ObjectEncodingCmd(c *Client, args []string) CommandResult {
	key := args[0]
	obj, e := c.db.lookupKey(key)
	if !e {
		return Result(Null())
	}
	return Result(BulkString(obj.encoding.StrRep()))
}
func ObjFreqCmd(c *Client, args []string) CommandResult {
	return Failed(Error("Not yet implemented"))

}

func ObjFreqIdleTime(c *Client, args []string) CommandResult {
	return Failed(Error("Not yet implemented"))

}

func ObjRefcount(c *Client, args []string) CommandResult {
	return Failed(Error("Not yet implemented"))

}

// ! async or sync here...!
func FlushAll(c *Client, args []string) CommandResult {
	// delete all keys
	c.server.db.Flush()
	return Result(SimpleString("OK"))

}
