package redis

import (
	"sync"
	"time"
)

type RedisDb struct {
	mu   sync.RWMutex
	dict map[string]*RedisObject
} // todo: versions

func NewDb() *RedisDb {
	return &RedisDb{dict: map[string]*RedisObject{}}
}

func (db *RedisDb) lookupKey(key string) (*RedisObject, bool) {
	obj := db.dict[key]
	if obj == nil {
		return nil, false
	}

	if obj.expired(time.Now()) {
		delete(db.dict, key)
		return nil, false
	}

	return obj, true
}
func (db *RedisDb) setKey(key string, obj *RedisObject) {
	db.dict[key] = obj
}
func (db *RedisDb) deleteKey(key string) bool {

	obj := db.dict[key]

	if obj == nil {
		return false
	}

	delete(db.dict, key)
	return true
}
func (o *RedisObject) expired(now time.Time) bool {
	return !o.expiresAt.IsZero() && now.After(o.expiresAt)

}

func (obj *RedisObject) ttlForObject(now time.Time) int {
	if obj.expiresAt.IsZero() {
		return -1
	}

	if obj.expired(now) {
		return -2
	}

	return int(time.Until(obj.expiresAt).Seconds())
}
