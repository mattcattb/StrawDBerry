package redis

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

	switch ObjectType(obj.typ) {
	case StringObject:
		return Result(SimpleString("string"))

	case ListObject:
		return Result(SimpleString("list"))

	case SetObject:
		return Result(SimpleString("set"))
	case ZSetObject:
		return Result(SimpleString("zset"))

	case HashObject:
		return Result(SimpleString("hash"))
	}

	return Failed(Error("Unknown type"))
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

func ObjCommand(c *Client, args []string) CommandResult {
	cmdType, key := args[0], args[1]

	obj, e := c.db.lookupKey(key)

	if !e {
		return Result(Null())
	}

	switch cmdType {
	case "ENCODING":
		enc := obj.encoding
		return Result(BulkString(enc.StrRep()))
	case "FREQ":
	case "IDLETIME":
		// time in seconds since the last access to the value stored at key.

	case "REFCOUNT":
		// reference count of the stored at key... ?
	}

	return Failed(wrongArgs("COMMAND"))
}
