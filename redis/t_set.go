package redis

func registerSetCSpec(sh *SpecHandler) {

	setSpecs := map[string]CommandSpec{

		"SADD": {
			handler: SAdd,
			group:   SetGroup,
			flags:   CmdWrite,
			arity:   -2,
		},
		"SCARD": {
			handler: SCard,
			group:   SetGroup,
			flags:   CmdRead,
			arity:   1,
		},
		"SREM": {
			handler: SRem,
			group:   SetGroup,
			flags:   CmdWrite,
			arity:   -2,
		},
		"SISMEM": {
			handler: SIsMem,
			group:   SetGroup,
			flags:   CmdWrite,
			arity:   2,
		},
	}

	sh.registerCommandSpecs(setSpecs)

}

func newSetObj() *RedisObject {
	return &RedisObject{
		typ:      SetObject,
		encoding: EncodingSetMap,
		ptr:      setMapPayload{},
	}
}

func setObjValue(obj *RedisObject) (setMapPayload, error) {

	if obj.typ != SetObject {
		return nil, ErrWrongType
	}

	switch obj.encoding {
	case EncodingSetMap:
		val, ok := obj.ptr.(setMapPayload)
		if !ok {
			return nil, ErrInvalidEncoding
		}

		return val, nil
	}

	return nil, ErrInvalidEncoding

}

func setTypeAdd(obj *RedisObject, members ...string) (int, error) {
	set, err := setObjValue(obj)
	if err != nil {
		return 0, err
	}

	added := 0

	for _, member := range members {
		if _, exists := set[member]; exists {
			continue
		}

		set[member] = struct{}{}
		added++
	}
	return added, nil
}

func setTypeExists(obj *RedisObject, member string) (bool, error) {
	set, err := setObjValue(obj)
	if err != nil {
		return false, err
	}

	_, exists := set[member]
	return exists, nil
}

func setTypeDel(obj *RedisObject, member string) (bool, error) {
	set, err := setObjValue(obj)
	if err != nil {
		return false, err
	}
	_, exists := set[member]

	delete(set, member)

	return exists, nil
}

func SAdd(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	obj, exists := c.db.lookupKey(key)

	if !exists {
		obj = newSetObj()
		c.db.setKey(key, obj)
	}

	added, err := setTypeAdd(obj, members...)
	c.server.dirty += uint64(added)

	if err != nil {
		return Failed(wrongTypeError())
	}

	return CommandResult{Reply: Integer(added)}
}

func SCard(c *Client, args []string) CommandResult {
	// Returns the set cardinality (number of elements) of the set stored at key.
	key := args[0]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(0))
	}
	m, err := setObjValue(obj)

	if err != nil {
		return Failed(wrongTypeError())
	}

	return Result(Integer(len(m)))
}

func SRem(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
}

func SIsMem(c *Client, args []string) CommandResult {
	return Failed(Error("ERR not implemented"))
}
