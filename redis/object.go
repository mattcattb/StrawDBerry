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
	EncodingHashMap
	EncodingSetMap
)

type objectPayload interface {
	objectPayload()
}

type rawStringPayload string
type intStringPayload int
type hashMapPayload map[string]string
type setMapPayload map[string]struct{}

func (rawStringPayload) objectPayload() {}
func (intStringPayload) objectPayload() {}
func (hashMapPayload) objectPayload()   {}
func (setMapPayload) objectPayload()    {}

type RedisObject struct {
	typ       ObjectType
	encoding  ObjectEncoding
	ptr       objectPayload
	expiresAt time.Time
}

func newObject(typ ObjectType, encoding ObjectEncoding, ptr objectPayload) *RedisObject {
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
