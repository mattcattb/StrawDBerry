package redis

import "maps"

func newHashObject() *RedisObject {
	return newObject(
		ObjectTypeHash,
		ObjectEncodingHashMap,
		make(hashMapPayload),
	)
}

func hashMapFromObject(obj *RedisObject) (hashMapPayload, error) {
	if err := obj.checkType(ObjectTypeHash); err != nil {
		return nil, err
	}

	if obj.encoding != ObjectEncodingHashMap {
		return nil, ErrInvalidEncoding
	}

	hash, ok := obj.payload.(hashMapPayload)
	if !ok {
		return nil, ErrInvalidEncoding
	}

	return hash, nil
}

func cloneHashPayload(obj *RedisObject) (objectPayload, error) {
	hash, err := hashMapFromObject(obj)
	if err != nil {
		return nil, err
	}

	return maps.Clone(hash), nil
}

type hashSetResult struct {
	added   bool
	changed bool
}

func hashSet(obj *RedisObject, field, value string) (hashSetResult, error) {

	hash, err := hashMapFromObject(obj)
	if err != nil {
		return hashSetResult{}, err
	}

	oldValue, exists := hash[field]
	hash[field] = value

	return hashSetResult{
		added:   !exists,
		changed: !exists || oldValue != value,
	}, nil
}

func hashGet(obj *RedisObject, field string) (string, bool, error) {
	hash, err := hashMapFromObject(obj)
	if err != nil {
		return "", false, err
	}

	val, exists := hash[field]

	return val, exists, nil
}

func hashDelete(obj *RedisObject, fields ...string) (deleted int, err error) {
	hash, err := hashMapFromObject(obj)
	if err != nil {
		return 0, err
	}

	for _, field := range fields {
		if _, exists := hash[field]; !exists {
			continue
		}
		delete(hash, field)
		deleted++
	}

	return deleted, nil
}

func hashExists(obj *RedisObject, field string) (exists bool, err error) {
	hash, err := hashMapFromObject(obj)

	if err != nil {
		return false, err
	}
	_, ok := hash[field]

	return ok, nil

}

func hashCardinality(obj *RedisObject) (int, error) {
	hash, err := hashMapFromObject(obj)
	if err != nil {
		return 0, err
	}
	return len(hash), nil
}

func hashEntries(obj *RedisObject) ([][2]string, error) {
	hash, err := hashMapFromObject(obj)
	if err != nil {
		return nil, err
	}

	entries := make([][2]string, 0, len(hash))
	for field, value := range hash {
		entries = append(entries, [2]string{field, value})
	}
	return entries, nil
}

func HGet(c *Client, args []string) CommandResult {
	// HGET key field
	// returns Array field, value
	key, field := args[0], args[1]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists, _ := c.db.lookupKeyLocked(key)

	if !exists {
		return Result(Null())
	}

	value, found, err := hashGet(obj, field)
	if err != nil {
		return commandFailure(err)
	}

	if !found {
		return Result(Null())
	}

	return Result(BulkString(value))

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
	changed := false

	obj, exists, _ := c.db.lookupKeyLocked(key)

	if !exists {
		obj = newHashObject()
	}

	for i := 0; i < len(kvArray); i += 2 {
		field, value := kvArray[i], kvArray[i+1]

		result, err := hashSet(obj, field, value)
		if err != nil {
			return commandFailure(err)
		}
		if result.added {
			setCount += 1
		}
		if result.changed {
			changed = true
		}
	}

	c.db.setKeyLocked(key, obj)
	if changed {
		c.server.dirty += 1
	}
	return Result(Integer(setCount))

}

func HDel(c *Client, args []string) CommandResult {
	// key field [field ...]
	key, fieldValues := args[0], args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists, _ := c.db.lookupKeyLocked(key)

	if !exists {
		return Result(Integer(0))
	}

	delCount, err := hashDelete(obj, fieldValues...)
	if err != nil {
		return commandFailure(err)
	}

	if delCount > 0 {
		c.server.dirty += 1
		remaining, err := hashCardinality(obj)
		if err != nil {
			return commandFailure(err)
		}
		if remaining == 0 {
			delete(c.db.dict, key)
			c.db.stats.deletedKeys++
		}
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

	obj, exists, _ := c.db.lookupKeyLocked(key)

	if !exists {
		return Result(Array([]Value{}))
	}

	entries, err := hashEntries(obj)
	if err != nil {
		return commandFailure(err)
	}

	returnValues := make([]Value, 0)

	for _, entry := range entries {
		returnValues = append(returnValues, BulkString(entry[0]), BulkString(entry[1]))
	}

	return Result(Array(returnValues))

}

func HExists(c *Client, args []string) CommandResult {
	// HEXISTS key field
	key, field := args[0], args[1]

	obj, exists, _ := c.db.lookupKey(key)

	if !exists {
		return Result(Integer(0))
	}

	found, err := hashExists(obj, field)
	if err != nil {
		return commandFailure(err)
	}

	if found {
		return Result(Integer(1))
	}

	return Result(Integer(0))

}
