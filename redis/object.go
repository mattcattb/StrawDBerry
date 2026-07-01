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
