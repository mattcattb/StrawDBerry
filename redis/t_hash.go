package redis

func registerTHashCommandSpec(sh *SpecHandler) {

	hSpecs := map[string]CommandSpec{
		"HSET": {
			arity:   -3,
			flags:   CmdWrite,
			handler: HSet,
		},

		"HDEL": {
			arity:   2,
			flags:   CmdWrite,
			handler: HDel},
		"HGETALL": {
			arity:   1,
			flags:   CmdRead,
			handler: HGetAll,
		},
		"HEXISTS": {
			arity:   2,
			flags:   CmdRead,
			handler: HExists,
		},
	}

	for k, v := range hSpecs {
		v.group = HashGroup
		v.name = k
		hSpecs[k] = v
	}
	sh.registerCommandSpecs(hSpecs)

}

func newHashRObject() *RedisObject {
	return &RedisObject{
		typ:      HashObject,
		encoding: EncodingHashMap,
		ptr:      hashMapPayload(map[string]string{}),
	}
}

func hashObjValue(obj *RedisObject) (map[string]string, error) {

	var newMap map[string]string
	if obj.typ != HashObject {
		return newMap, ErrWrongType
	}

	switch obj.encoding {
	case EncodingHashMap:
		val, ok := obj.ptr.(hashMapPayload)
		if !ok {
			return newMap, ErrInvalidEncoding
		}

		return val, nil
	}

	return newMap, ErrInvalidEncoding
}

/*
func hashTypeSet(obj *RedisObject, key, value string) (bool, error) {
	obj.encoding = EncodingMap

	hash, err := hashObjValue(obj)
	if err != nil {
		return false, err
	}

}

func hashTypeGet(obj *RedisObject, field string) (string, bool, error)
func hashTypeDel(obj *RedisObject, fields ...string) (deleted int, err error) {

} */

func HGet(c *Client, args []string) CommandResult {
	// HGET key field
	// returns Array field, value
	key, field := args[0], args[1]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Null())
	}

	hash, err := hashObjValue(obj)
	if err != nil {
		return Failed(wrongTypeError())
	}

	val, exists := hash[field]
	if !exists {
		return Result(Null())
	}

	return Result(BulkString(val))

}

func HSet(c *Client, args []string) CommandResult {
	// HSET key field value [field value ...]
	// returns int of set fields
	key := args[0]
	kvArray := args[1:]

	if len(kvArray) == 0 || len(kvArray)%2 != 0 {
		return Failed(wrongArgs("HSET"))
	}
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	setCount := 0

	obj, exists := c.db.lookupKey(key)

	if !exists {
		obj = newHashRObject()
	}

	hashObj, err := hashObjValue(obj)
	if err != nil {
		return Failed(wrongTypeError())
	}

	for i := 0; i < len(kvArray); i += 2 {
		field, value := kvArray[i], kvArray[i+1]

		hashObj[field] = value
		setCount += 1
	}

	obj.ptr = hashMapPayload(hashObj)

	c.db.setKey(key, obj)
	c.server.dirty += 1
	return Result(Integer(setCount))

}

func HDel(c *Client, args []string) CommandResult {
	// key field [field ...]
	key, fieldValues := args[0], args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(0))
	}

	hash, err := hashObjValue(obj)

	if err != nil {
		return Failed(wrongTypeError())
	}

	delCount := 0
	for _, delField := range fieldValues {
		_, exists := hash[delField]
		if exists {
			delete(hash, delField)
			delCount += 1
		}
	}

	if delCount > 0 {
		c.server.dirty += 1
	}

	return Result(Integer(delCount))

}
func HGetAll(c *Client, args []string) CommandResult {
	// HGETALL key
	// returns Array field, value

	//  a list of fields and their values, or an empty list when key does not exist

	key := args[0]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Array([]Value{}))
	}

	hash, err := hashObjValue(obj)
	if err != nil {
		return Failed(wrongTypeError())
	}

	returnValues := make([]Value, 0)

	for field, val := range hash {
		returnValues = append(returnValues, BulkString(field), BulkString(val))
	}

	return Result(Array(returnValues))

}

func HExists(c *Client, args []string) CommandResult {
	// HEXISTS key field
	key, field := args[0], args[1]

	obj, exists := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(0))
	}

	hash, err := hashObjValue(obj)

	if err != nil {
		return Failed(wrongTypeError())
	}

	_, exists = hash[field]

	if exists {
		return Result(Integer(1))
	}

	return Result(Integer(0))

}
