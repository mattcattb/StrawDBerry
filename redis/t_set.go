package redis

import "sort"

func newSetObj() *RedisObject {
	return &RedisObject{
		typ:       SetObject,
		encoding:  EncodingSetMap,
		ptr:       setMapPayload{},
		expiresAt: noExpiration,
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
	obj, exists := c.db.lookupKeyLocked(key)

	if !exists {
		obj = newSetObj()
		c.db.setKeyLocked(key, obj)
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

	obj, exists := c.db.lookupKeyLocked(key)

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

	key, members := args[0], args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	rObj, exists := c.db.lookupKeyLocked(key)

	if !exists {
		return Result(Integer(0))
	}

	n := 0

	for _, m := range members {
		del, err := setTypeDel(rObj, m)
		if err != nil {
			return Failed(wrongTypeError())
		}

		if del {
			n += 1
		}
	}

	if n > 0 {
		c.server.dirty += uint64(n)
		set, _ := setObjValue(rObj)
		if len(set) == 0 {
			delete(c.db.dict, key)
			c.db.stats.deletedKeys++
		}
	}

	return Result(Integer(n))
}

func SMIsMem(c *Client, args []string) CommandResult {

	key, members := args[0], args[1:]

	respValues := make([]Value, len(members))

	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	rObj, e := c.db.lookupKeyLocked(key)
	for i, _ := range respValues {
		respValues[i] = Integer(0)
	}

	if !e {
		return Result(Array(respValues))
	}

	for i, m := range members {

		exists, err := setTypeExists(rObj, m)

		if err != nil {
			return Failed(wrongTypeError())
		}
		n := 0
		if exists {
			n++
		}

		respValues[i] = Integer(n)

	}
	return Result(Array(respValues))
}

func SIsMem(c *Client, args []string) CommandResult {
	key, member := args[0], args[1]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()
	rObj, e := c.db.lookupKeyLocked(key)

	n := 0
	if !e {
		return Result(Integer(n))
	}

	exists, err := setTypeExists(rObj, member)

	if err != nil {
		return Failed(wrongTypeError())
	}

	if exists {
		n++
	}
	return Result(Integer(n))
}

func setValuesForKeysLocked(c *Client, keys []string) ([]setMapPayload, error) {
	sets := make([]setMapPayload, len(keys))
	for i, key := range keys {
		obj, exists := c.db.lookupKeyLocked(key)
		if !exists {
			continue
		}
		set, err := setObjValue(obj)
		if err != nil {
			return nil, err
		}
		sets[i] = set
	}
	return sets, nil
}

func setMembersReply(members map[string]struct{}) CommandResult {
	ordered := make([]string, 0, len(members))
	for member := range members {
		ordered = append(ordered, member)
	}
	sort.Strings(ordered)

	values := make([]Value, len(ordered))
	for i, member := range ordered {
		values[i] = BulkString(member)
	}
	return Result(Array(values))
}

func SMembers(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	sets, err := setValuesForKeysLocked(c, args)
	if err != nil {
		return Failed(wrongTypeError())
	}
	return setMembersReply(sets[0])
}

func SDiff(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	sets, err := setValuesForKeysLocked(c, args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	remaining := make(map[string]struct{}, len(sets[0]))
	for member := range sets[0] {
		remaining[member] = struct{}{}
	}
	for _, set := range sets[1:] {
		for member := range set {
			delete(remaining, member)
		}
	}
	return setMembersReply(remaining)
}

func SInter(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	sets, err := setValuesForKeysLocked(c, args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	intersection := make(map[string]struct{})
	for member := range sets[0] {
		presentInAll := true
		for _, set := range sets[1:] {
			if _, exists := set[member]; !exists {
				presentInAll = false
				break
			}
		}
		if presentInAll {
			intersection[member] = struct{}{}
		}
	}
	return setMembersReply(intersection)
}

func SUnion(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	sets, err := setValuesForKeysLocked(c, args)
	if err != nil {
		return Failed(wrongTypeError())
	}

	union := make(map[string]struct{})
	for _, set := range sets {
		for member := range set {
			union[member] = struct{}{}
		}
	}
	return setMembersReply(union)
}
