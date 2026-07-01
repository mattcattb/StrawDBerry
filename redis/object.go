package redis

import (
	"time"
)

type ObjectType uint8

const (
	StringObject ObjectType = iota
	ListObject
	SetObject
	ZSetObject
	HashObject
)

type RedisObject struct {
	typ       ObjectType
	encoding  ObjectEncoding
	ptr       objectPayload
	expiresAt int64
}

const noExpiration int64 = -1

func (o *RedisObject) setExprMs(durMs int64) {
	expiresAt := noExpiration
	if durMs > 0 {
		expiresAt = time.Now().UnixMilli() + durMs
	}

	o.expiresAt = expiresAt
}

func (o *RedisObject) copy() RedisObject {
	return RedisObject{
		typ:       o.typ,
		encoding:  o.encoding,
		ptr:       o.ptr,
		expiresAt: o.expiresAt,
	}
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

func newObject(typ ObjectType, encoding ObjectEncoding, ptr objectPayload) *RedisObject {
	return &RedisObject{
		typ:       typ,
		encoding:  encoding,
		ptr:       ptr,
		expiresAt: noExpiration,
	}
}

func checkObjectType(obj *RedisObject, typ ObjectType) error {
	if obj == nil {
		return ErrWrongType
	}
	if obj.typ != typ {
		return ErrWrongType
	}

	return nil
}

func serializeObj(obj *RedisObject) (string, error)

func restoreObj(hash string) (RedisObject, error)
