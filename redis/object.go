package redis

import (
	"time"
)

type ObjectType uint8

const (
	ObjectTypeString ObjectType = iota
	ObjectTypeList
	ObjectTypeSet
	ObjectTypeZSet
	ObjectTypeHash
)

func (o ObjectType) String() string {
	switch ObjectType(o) {
	case ObjectTypeString:
		return "string"
	case ObjectTypeList:
		return "list"
	case ObjectTypeSet:
		return "set"
	case ObjectTypeZSet:
		return "zset"
	case ObjectTypeHash:
		return "hash"
	default:
		return "UNKNOWN"
	}

}

type RedisObject struct {
	typ       ObjectType
	encoding  ObjectEncoding
	payload   objectPayload
	expiresAt int64
}

const noExpiration int64 = -1

func (o *RedisObject) persist() {
	o.expiresAt = noExpiration
}

func (o *RedisObject) expired() bool {
	now := time.Now().UnixMilli()
	return o.expiresAt != noExpiration && o.expiresAt <= now
}

func (obj *RedisObject) ttlSeconds() int64 {
	if obj.expiresAt == noExpiration {
		return -1
	}

	now := time.Now().UnixMilli()
	if obj.expiresAt <= now {
		return -2
	}

	return (obj.expiresAt - now) / 1000
}

func newObject(typ ObjectType, encoding ObjectEncoding, payload objectPayload) *RedisObject {
	return &RedisObject{
		typ:       typ,
		encoding:  encoding,
		payload:   payload,
		expiresAt: noExpiration,
	}
}

func (o *RedisObject) clone() (*RedisObject, error) {

	var (
		payload objectPayload
		err     error
	)

	switch o.typ {
	case ObjectTypeHash:
		payload, err = cloneHashPayload(o)
	case ObjectTypeSet:
		payload, err = cloneSetPayload(o)

	case ObjectTypeZSet:
		payload, err = cloneZSetPayload(o)
	case ObjectTypeString:
		payload, err = cloneStringPayload(o)
	default:
		return nil, ErrInvalidObjectType
	}

	if err != nil {
		return nil, err
	}

	return &RedisObject{
		typ:       o.typ,
		encoding:  o.encoding,
		payload:   payload,
		expiresAt: o.expiresAt,
	}, nil
}

func (o *RedisObject) checkType(objType ObjectType) error {

	if o == nil {
		return ErrWrongType
	}
	if o.typ == objType {
		return nil
	}

	return ErrWrongType
}
