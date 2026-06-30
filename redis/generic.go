package redis

import (
	"time"
)

func registerGenericCommands(sh *SpecHandler) {
	sh.registerCommandSpecs(map[string]CommandSpec{
		"EXISTS": {
			handler: Exists,
			arity:   -1,
			group:   GenericGroup,
			flags:   CmdRead,
		},
		"TYPE": {
			arity:   1,
			group:   GenericGroup,
			flags:   CmdRead,
			handler: Type,
		},
		"TTL": {
			arity:   1,
			group:   GenericGroup,
			flags:   CmdWrite,
			handler: Ttl,
		},
		"DEL": {
			arity:   -1,
			group:   GenericGroup,
			flags:   CmdWrite,
			handler: Del,
		},
		"COPY": {handler: Copy, arity: 2, group: GenericGroup, flags: CmdWrite},
	})

}

// OBJECT ENCODING

func Copy(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
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

	if obj.expiresAt.IsZero() {
		return Result(Integer(-1))
	}

	seconds := int(time.Until(obj.expiresAt).Seconds())
	if seconds < 0 {
		delete(c.db.dict, key)
		c.server.dirty += 1
		return Result(Integer(-2))
	}

	return Result(Integer(seconds))
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
