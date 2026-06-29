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

type ObjectEncoding uint8

const (
	EncodingRaw ObjectEncoding = iota
	EncodingInt
	EncodingMap
	EncodingListpack
	EncodingIntSet
	EncodingQuicklist
)

type RedisObject struct {
	typ       ObjectType
	encoding  ObjectEncoding
	ptr       any
	expiresAt time.Time
}

func newObject(typ ObjectType, encoding ObjectEncoding, ptr any) *RedisObject {
	return &RedisObject{
		typ:      typ,
		encoding: encoding,
		ptr:      ptr,
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

/*
func newStringObject(value string) *RedisObject {
	if n, ok := tryParseInt(value); ok {
		return &RedisObject{
			typ:      StringObject,
			encoding: EncodingInt,
			ptr:      n,
		}
	}
	return &RedisObject{
		typ:      StringObject,
		encoding: EncodingRaw,
		ptr:      value,
	}

}
*/
