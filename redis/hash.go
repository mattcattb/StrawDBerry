package redis

import "go-redis/resp"

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

func (ce *CommandExecutor) HGet(args []resp.Value) resp.Value {
	// HGET key field
	// returns resp.Array field, value
	if len(args) != 2 {
		return wrongArgs("hget")
	}
	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil {
		return wrongTypeError()
	}

	key, field := strArgs[0], strArgs[1]

	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Null()
	}

	hash, err := hashObjValue(obj)
	if err != nil {
		return wrongTypeError()
	}

	val, exists := hash[field]
	if !exists {
		return resp.Null()
	}

	return resp.BulkString(val)

}

func (ce *CommandExecutor) HSet(args []resp.Value) resp.Value {
	// HSET key field value [field value ...]
	// returns int of set fields

	strArgs, err := parseBulkRespStringCommands(args)

	if err != nil {
		return wrongTypeError()
	}

	if len(strArgs) < 3 {
		return wrongArgs("hset")
	}

	key := strArgs[0]
	kvArray := strArgs[1:]

	if len(kvArray) == 0 || len(kvArray)%2 != 0 {
		return wrongArgs("HSET")
	}
	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	setCount := 0

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		obj = newHashRObject()
	}

	hashObj, err := hashObjValue(obj)
	if err != nil {
		return wrongTypeError()
	}

	for i := 0; i < len(strArgs); i += 2 {
		field, value := strArgs[i], strArgs[i+1]

		hashObj[field] = value
		setCount += 1
	}

	return resp.Integer(setCount)

}

func (ce *CommandExecutor) HDel(args []resp.Value) resp.Value {
	// key field [field ...]
	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil || len(strArgs) < 2 {
		return wrongArgs("HDel")
	}
	key, fieldValues := strArgs[0], strArgs[1:]

	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Integer(0)
	}

	hash, err := hashObjValue(obj)

	if err != nil {
		return resp.Integer(0)
	}

	delCount := 0
	for _, delField := range fieldValues {
		delete(hash, delField)
		delCount += 1
	}

	return resp.Integer(delCount)

}
func (ce *CommandExecutor) HGetAll(args []resp.Value) resp.Value {
	// HGETALL key
	// returns resp.Array field, value

	//  a list of fields and their values, or an empty list when key does not exist

	if len(args) != 1 {
		return wrongArgs("hgetall")
	}

	key, ok := args[0].BulkString()

	if !ok {
		return syntaxError()
	}

	ce.db.mu.Lock()
	defer ce.db.mu.Unlock()

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Array([]resp.Value{})
	}

	hash, err := hashObjValue(obj)
	if err != nil {
		return wrongTypeError()
	}

	returnValues := make([]resp.Value, 0)

	for _, k := range hash {
		val, _ := hash[k]

		returnValues = append(returnValues, resp.BulkString(val))
	}

	return resp.Array(returnValues)

}

func (ce *CommandExecutor) HExists(args []resp.Value) resp.Value {
	// HEXISTS key field

	if len(args) != 2 {
		return wrongArgs("hexists")
	}

	strArgs, err := parseBulkRespStringCommands(args)
	if err != nil {
		return wrongTypeError()
	}

	key, field := strArgs[0], strArgs[1]

	obj, exists := ce.db.lookupKey(key)

	if !exists {
		return resp.Integer(0)
	}

	hash, err := hashObjValue(obj)

	if err != nil {
		return wrongTypeError()
	}

	_, exists = hash[field]

	if exists {
		return resp.Integer(1)
	}

	return resp.Integer(0)

}
