package store

import (
	"errors"
	"sync"
	"time"
)

type ObjectType string

const (
	StringObject ObjectType = "string"
	HashObject   ObjectType = "hash"
	SetObject    ObjectType = "set"
	ListObject   ObjectType = "list"
)

type RedisObject struct {
	Type      ObjectType
	Value     any
	ExpiresAt time.Time
}

type RedisStore struct {
	mu sync.RWMutex
	db map[string]RedisObject
}

var ErrWrongType = errors.New("wrong type")

func NewStore() {

}
