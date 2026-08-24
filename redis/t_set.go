package redis

import (
	"maps"
	"sort"
)

func newSetObject() *RedisObject {
	return newObject(
		ObjectTypeSet,
		ObjectEncodingSetMap,
		make(setMapPayload),
	)
}

func cloneSetPayload(obj *RedisObject) (objectPayload, error) {

	set, err := setMapFromObject(obj)

	if err != nil {
		return nil, err
	}

	return maps.Clone(set), nil

}

// setMapFromObject returns the map representation of a set. Semantic set
// operations should use the set behavior helpers instead so callers do not depend
// on a particular encoding.
func setMapFromObject(obj *RedisObject) (setMapPayload, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return nil, err
	}
	if obj.encoding != ObjectEncodingSetMap {
		return nil, ErrInvalidEncoding
	}

	set, ok := obj.payload.(setMapPayload)
	if !ok {
		return nil, ErrInvalidEncoding
	}
	return set, nil
}

func setCardinality(obj *RedisObject) (int, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case ObjectEncodingSetMap:
		set, err := setMapFromObject(obj)
		if err != nil {
			return 0, err
		}
		return len(set), nil
	default:
		return 0, ErrInvalidEncoding
	}
}

func setAdd(obj *RedisObject, members ...string) (int, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case ObjectEncodingSetMap:
		set, err := setMapFromObject(obj)
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

func setRemove(obj *RedisObject, members ...string) (int, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return 0, err
	}

	switch obj.encoding {
	case ObjectEncodingSetMap:
		set, err := setMapFromObject(obj)
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

func setContains(obj *RedisObject, member string) (bool, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return false, err
	}

	switch obj.encoding {
	case ObjectEncodingSetMap:
		set, err := setMapFromObject(obj)
		if err != nil {
			return false, err
		}
		_, exists := set[member]
		return exists, nil
	default:
		return false, ErrInvalidEncoding
	}
}

func setMembers(obj *RedisObject) ([]string, error) {
	if err := obj.checkType(ObjectTypeSet); err != nil {
		return nil, err
	}

	switch obj.encoding {
	case ObjectEncodingSetMap:
		set, err := setMapFromObject(obj)
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

func setDiff(first *RedisObject, rest ...*RedisObject) ([]string, error) {
	if first != nil {
		if _, err := setCardinality(first); err != nil {
			return nil, err
		}
	}
	for _, obj := range rest {
		if obj != nil {
			if _, err := setCardinality(obj); err != nil {
				return nil, err
			}
		}
	}
	if first == nil {
		return []string{}, nil
	}

	candidates, err := setMembers(first)
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
			exists, err := setContains(obj, member)
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

func setInter(objects ...*RedisObject) ([]string, error) {
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

		length, err := setCardinality(obj)
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

	candidates, err := setMembers(objects[smallestIndex])
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
			exists, err := setContains(obj, member)
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

func setUnion(objects ...*RedisObject) ([]string, error) {
	unique := make(map[string]struct{})

	for _, obj := range objects {
		if obj == nil {
			continue
		}
		members, err := setMembers(obj)
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

func SAdd(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		obj = newSetObject()
	}

	added, err := setAdd(obj, members...)
	if err != nil {
		return commandFailure(err)
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

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	length, err := setCardinality(obj)
	if err != nil {
		return commandFailure(err)
	}
	return Result(Integer(length))
}

func SRem(c *Client, args []string) CommandResult {
	key := args[0]
	members := args[1:]

	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	removed, err := setRemove(obj, members...)
	if err != nil {
		return commandFailure(err)
	}
	if removed == 0 {
		return Result(Integer(0))
	}

	c.server.dirty += uint64(removed)
	remaining, err := setCardinality(obj)
	if err != nil {
		return commandFailure(err)
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

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		for i := range replies {
			replies[i] = Integer(0)
		}
		return Result(Array(replies))
	}

	for i, member := range members {
		exists, err := setContains(obj, member)
		if err != nil {
			return commandFailure(err)
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

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		return Result(Integer(0))
	}

	exists, err := setContains(obj, member)
	if err != nil {
		return commandFailure(err)
	}
	if exists {
		return Result(Integer(1))
	}
	return Result(Integer(0))
}

func lookupSetObjectsLocked(db *RedisDb, keys []string) ([]*RedisObject, error) {
	objects := make([]*RedisObject, len(keys))
	for i, key := range keys {
		obj, exists, _ := db.lookupKeyLocked(key)
		if !exists {
			continue
		}
		if err := obj.checkType(ObjectTypeSet); err != nil {
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

	obj, exists, _ := c.db.lookupKeyLocked(key)
	if !exists {
		return setMembersReply([]string{})
	}
	members, err := setMembers(obj)
	if err != nil {
		return commandFailure(err)
	}
	return setMembersReply(members)
}

func SDiff(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return commandFailure(err)
	}
	members, err := setDiff(objects[0], objects[1:]...)
	if err != nil {
		return commandFailure(err)
	}
	return setMembersReply(members)
}

func SInter(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return commandFailure(err)
	}
	members, err := setInter(objects...)
	if err != nil {
		return commandFailure(err)
	}
	return setMembersReply(members)
}

func SUnion(c *Client, args []string) CommandResult {
	c.db.mu.Lock()
	defer c.db.mu.Unlock()

	objects, err := lookupSetObjectsLocked(c.db, args)
	if err != nil {
		return commandFailure(err)
	}
	members, err := setUnion(objects...)
	if err != nil {
		return commandFailure(err)
	}
	return setMembersReply(members)
}
