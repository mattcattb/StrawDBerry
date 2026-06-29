package redis

func registerTHashCommandSpec(sh *SpecHandler) {

	hSpecs := map[string]CommandSpec{
		"HSET": {
			minArgs: 3,
			maxArgs: -1,
			flags:   CmdWrite,
			handler: HSet,
		},

		"HDEL": {minArgs: 2,
			maxArgs: -1,
			flags:   CmdWrite,
			handler: HDel},
		"HGETALL": {
			minArgs: 1,
			maxArgs: 1,
			flags:   CmdRead,
			handler: HGetAll,
		},
		"HEXISTS": {minArgs: 2,
			maxArgs: 2,
			flags:   CmdRead,
			handler: HExists},
	}

	sh.registerCommandSpecs(hSpecs)

	for k, v := range hSpecs {
		v.group = HashGroup
		v.name = k
		hSpecs[k] = v
	}

}

func newHashRObject() *RedisObject {
	return &RedisObject{
		typ:      HashObject,
		encoding: EncodingMap,
		ptr:      map[string]string{},
	}
}

func hashObjValue(obj *RedisObject) (map[string]string, error) {

	var newMap map[string]string
	if obj.typ != HashObject {
		return newMap, ErrWrongType
	}

	switch obj.encoding {
	case EncodingMap:
		val, ok := obj.ptr.(map[string]string)
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

func HGet(c *Client, args []Value) CommandResult {
	// HGET key field
	// returns Array field, value
	if len(args) != 2 {
		return Failed(wrongArgs("hget"))
	}
	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	key, field := strArgs[0], strArgs[1]

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

func HSet(c *Client, args []Value) CommandResult {
	// HSET key field value [field value ...]
	// returns int of set fields

	strArgs, err := parseBulkRespStringCommands(args)

	if err != nil {
		return Failed(wrongTypeError())
	}

	key := strArgs[0]
	kvArray := strArgs[1:]

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

	obj.ptr = hashObj

	c.db.setKey(key, obj)
	c.server.dirty += 1
	return Result(Integer(setCount))

}

func HDel(c *CommandContext, args []Value) CommandResult {
	// key field [field ...]
	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil || len(strArgs) < 2 {
		return Failed(wrongArgs("HDel"))
	}
	key, fieldValues := strArgs[0], strArgs[1:]

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
		delete(hash, delField)
		delCount += 1
	}

	if delCount > 0 {
		c.server.dirty += 1
	}

	return Result(Integer(delCount))

}
func HGetAll(c *CommandContext, args []Value) CommandResult {
	// HGETALL key
	// returns Array field, value

	//  a list of fields and their values, or an empty list when key does not exist

	if len(args) != 1 {
		return Failed(wrongArgs("hgetall"))
	}

	key, ok := args[0].BulkString()

	if !ok {
		return Failed(syntaxError())
	}

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

func HExists(c *CommandContext, args []Value) CommandResult {
	// HEXISTS key field

	if len(args) != 2 {
		return Failed(wrongArgs("hexists"))
	}

	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	key, field := strArgs[0], strArgs[1]

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
