package redis

import (
	"errors"
	"sort"
)

func newSetObj() *RedisObject {
	return &RedisObject{
		typ:       SetObject,
		encoding:  EncodingSetMap,
		ptr:       setMapPayload{},
		expiresAt: noExpiration,
	}
}

// setMapValue returns the map representation of a set. Semantic set
// operations should use the setType helpers instead so callers do not depend
// on a particular encoding.
func setMapValue(obj *RedisObject) (setMapPayload, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return nil, err
	}
	if obj.encoding != EncodingSetMap {
		return nil, ErrInvalidEncoding
	}

	set, ok := obj.ptr.(setMapPayload)
	if !ok {
		return nil, ErrInvalidEncoding
	}
	return set, nil
}

func setTypeCardinality(obj *RedisObject) (int, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case EncodingSetMap:
		set, err := setMapValue(obj)
		if err != nil {
			return 0, err
		}
		return len(set), nil
	default:
		return 0, ErrInvalidEncoding
	}
}

func setTypeAdd(obj *RedisObject, members ...string) (int, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case EncodingSetMap:
		set, err := setMapValue(obj)
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
	default:
		return 0, ErrInvalidEncoding
	}
}

func setTypeRemove(obj *RedisObject, members ...string) (int, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case EncodingSetMap:
		set, err := setMapValue(obj)
		if err != nil {
			return 0, err
		}

		removed := 0
		for _, member := range members {
			if _, exists := set[member]; !exists {
				continue
			}
			delete(set, member)
			removed++
		}
		return removed, nil
	default:
		return 0, ErrInvalidEncoding
	}
}

func setTypeContains(obj *RedisObject, member string) (bool, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return false, err
	}

	switch obj.encoding {
	case EncodingSetMap:
		set, err := setMapValue(obj)
		if err != nil {
			return false, err
		}
		_, exists := set[member]
		return exists, nil
	default:
		return false, ErrInvalidEncoding
	}
}

func setTypeMembers(obj *RedisObject) ([]string, error) {
	if err := checkObjectType(obj, SetObject); err != nil {
		return nil, err
	}

	switch obj.encoding {
	case EncodingSetMap:
		set, err := setMapValue(obj)
		if err != nil {
			return nil, err
		}

		members := make([]string, 0, len(set))
		for member := range set {
			members = append(members, member)
		}
		return members, nil
	default:
		return nil, ErrInvalidEncoding
	}
}

func setTypeDiff(first *RedisObject, rest ...*RedisObject) ([]string, error) {
	if first != nil {
		if _, err := setTypeCardinality(first); err != nil {
			return nil, err
		}
	}
	for _, obj := range rest {
		if obj != nil {
			if _, err := setTypeCardinality(obj); err != nil {
				return nil, err
			}
		}
	}
	if first == nil {
		return []string{}, nil
	}

	candidates, err := setTypeMembers(first)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(candidates))

	for _, member := range candidates {
		excluded := false
		for _, obj := range rest {
			if obj == nil {
				continue
			}
			exists, err := setTypeContains(obj, member)
			if err != nil {
				return nil, err
			}
			if exists {
				excluded = true
				break
			}
		}
		if !excluded {
			result = append(result, member)
		}
	}

	return result, nil
}

func setTypeInter(objects ...*RedisObject) ([]string, error) {
	if len(objects) == 0 {
		return []string{}, nil
	}

	smallestIndex := -1
	smallestLen := 0
	hasEmptySet := false
	for i, obj := range objects {
		if obj == nil {
			hasEmptySet = true
			continue
		}

		length, err := setTypeCardinality(obj)
		if err != nil {
			return nil, err
		}
		if smallestIndex == -1 || length < smallestLen {
			smallestIndex = i
			smallestLen = length
		}
	}
	if hasEmptySet {
		return []string{}, nil
	}

	candidates, err := setTypeMembers(objects[smallestIndex])
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(candidates))

	for _, member := range candidates {
		presentInAll := true
		for i, obj := range objects {
			if i == smallestIndex {
				continue
			}
			exists, err := setTypeContains(obj, member)
			if err != nil {
				return nil, err
			}
			if !exists {
				presentInAll = false
				break
			}
		}
		if presentInAll {
			result = append(result, member)
		}
	}

	return result, nil
}

func setTypeUnion(objects ...*RedisObject) ([]string, error) {
	unique := make(map[string]struct{})

	for _, obj := range objects {
		if obj == nil {
			continue
		}
		members, err := setTypeMembers(obj)
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			unique[member] = struct{}{}
		}
	}

	result := make([]string, 0, len(unique))
	for member := range unique {
		result = append(result, member)
	}
	return result, nil
}

func setCommandFailure(err error) CommandResult {
	if errors.Is(err, ErrWrongType) {
		return Failed(wrongTypeError())
	}
	return Failed(internalError())
}

func SAdd(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		obj = newSetObj()
	}

	added, err := setTypeAdd(obj, members...)
	if err != nil {
		return setCommandFailure(err)
	}
	if !exists {
		c.db.setKeyLocked(key, obj)
	}
	if added > 0 {
		c.server.dirty += uint64(added)
	}

	return Result(Integer(added))
}

func SCard(c *Client, args []string) CommandResult {
	key := args[0]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	length, err := setTypeCardinality(obj)
	if err != nil {
		return setCommandFailure(err)
	}
	return Result(Integer(length))
}

func SRem(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	removed, err := setTypeRemove(obj, members...)
	if err != nil {
		return setCommandFailure(err)
	}
	if removed == 0 {
		return Result(Integer(0))
	}

	c.server.dirty += uint64(removed)
	remaining, err := setTypeCardinality(obj)
	if err != nil {
		return setCommandFailure(err)
	}
	if remaining == 0 {
		delete(c.db.dict, key)
		c.db.stats.deletedKeys++
	}

	return Result(Integer(removed))
}

func SMIsMem(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]
	replies := make([]Value, len(members))

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		for i := range replies {
			replies[i] = Integer(0)
		}
		return Result(Array(replies))
	}

	for i, member := range members {
		exists, err := setTypeContains(obj, member)
		if err != nil {
			return setCommandFailure(err)
		}
		if exists {
			replies[i] = Integer(1)
		} else {
			replies[i] = Integer(0)
		}
	}
	return Result(Array(replies))
}

func SIsMem(c *Client, args []string) CommandResult {
	key := args[0]
	member := args[1]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	exists, err := setTypeContains(obj, member)
	if err != nil {
		return setCommandFailure(err)
	}
	if exists {
		return Result(Integer(1))
	}
	return Result(Integer(0))
}

func lookupSetObjectsLocked(db *RedisDb, keys []string) ([]*RedisObject, error) {
	objects := make([]*RedisObject, len(keys))
	for i, key := range keys {
		obj, exists := db.lookupKeyLocked(key)
		if !exists {
			continue
		}
		if err := checkObjectType(obj, SetObject); err != nil {
			return nil, err
		}
		objects[i] = obj
	}
	return objects, nil
}

func setMembersReply(members []string) CommandResult {
	sort.Strings(members)

	values := make([]Value, len(members))
	for i, member := range members {
		values[i] = BulkString(member)
	}
	return Result(Array(values))
}

func SMembers(c *Client, args []string) CommandResult {
	key := args[0]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists := c.db.lookupKeyLocked(key)
	if !exists {
		return setMembersReply([]string{})
	}
	members, err := setTypeMembers(obj)
	if err != nil {
		return setCommandFailure(err)
	}
	return setMembersReply(members)
}

func SDiff(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return setCommandFailure(err)
	}
	members, err := setTypeDiff(objects[0], objects[1:]...)
	if err != nil {
		return setCommandFailure(err)
	}
	return setMembersReply(members)
}

func SInter(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return setCommandFailure(err)
	}
	members, err := setTypeInter(objects...)
	if err != nil {
		return setCommandFailure(err)
	}
	return setMembersReply(members)
}

func SUnion(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return setCommandFailure(err)
	}
	members, err := setTypeUnion(objects...)
	if err != nil {
		return setCommandFailure(err)
	}
	return setMembersReply(members)
}
