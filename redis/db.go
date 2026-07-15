package redis

import (
	"sync"
)

type DbStats struct {
	ksMisses    uint64
	ksHits      uint64
	expiredKeys uint64
	deletedKeys uint64
}
type RedisDb struct {
	mu    sync.RWMutex
	dict  map[string]*RedisObject
	stats DbStats
} // todo: versions

type DbStatsSnapshot struct {
	DbStats
	keys    uint64
	expires uint64
}

func NewDb() *RedisDb {
	return &RedisDb{dict: map[string]*RedisObject{}, mu: sync.RWMutex{}}
}

// !

func (db *RedisDb) lookupKey(key string) (*RedisObject, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.lookupKeyLocked(key)
}

func (db *RedisDb) lookupKeyLocked(key string) (*RedisObject, bool) {
	obj := db.dict[key]
	if obj == nil {
		// key miss
		db.stats.ksMisses++
		return nil, false
	}

	if obj.expired() {
		// key expiration
		db.stats.expiredKeys++
		db.stats.ksMisses++
		delete(db.dict, key)
		return nil, false
	}

	db.stats.ksHits++
	return obj, true
}

// do a memory check here?
func (db *RedisDb) setKey(key string, obj *RedisObject) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.setKeyLocked(key, obj)
}

func (db *RedisDb) setKeyLocked(key string, obj *RedisObject) {
	db.dict[key] = obj
}
func (db *RedisDb) deleteKey(key string) bool {

	db.mu.Lock()
	defer db.mu.Unlock()
	return db.deleteKeyLocked(key)
}

func (db *RedisDb) deleteKeyLocked(key string) bool {
	obj := db.dict[key]

	if obj == nil {
		db.stats.ksMisses++
		return false
	}
	db.stats.ksHits++
	db.stats.deletedKeys++

	delete(db.dict, key)
	return true
}

func (db *RedisDb) StatsSnapshot() DbStatsSnapshot {

	db.mu.Lock()
	defer db.mu.Unlock()

	curStats := db.stats
	keys := len(db.dict)

	return DbStatsSnapshot{DbStats: curStats, keys: uint64(keys), expires: 0}

}

func (db *RedisDb) Flush() {
	// remove all keys

	db.mu.Lock()
	defer db.mu.Unlock()

	for k, _ := range db.dict {
		db.deleteKeyLocked(k)
	}
}
