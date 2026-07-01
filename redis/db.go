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

	if obj.expired() {
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
func (o *RedisObject) expired() bool {
	now := time.Now().UnixMilli()
	return o.expiresAt != noExpiration && o.expiresAt <= now
}

func (obj *RedisObject) ttlForObject() int64 {
	if obj.expiresAt == noExpiration {
		return -1
	}

	now := time.Now().UnixMilli()
	if obj.expiresAt <= now {
		return -2
	}

	return (obj.expiresAt - now) / 1000
}
