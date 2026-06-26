package store

import (
	"strconv"
	"time"
)

type SetOptions struct {
	TTL time.Duration // Set the specified expire time, in seconds (a positive integer).
	// EX      int           // Set the specified expire time, in seconds (a positive integer).
	// PX      int           // Set the specified expire time, in milliseconds (a positive integer).
	KeepTTL bool // Retain the time to live associated with the key
	NX      bool // Only set the key if it does not already exist.
	XX      bool // Only set the key if it already exists.
	Get     bool // Return the old string stored at the key, or nil if the key did not exist. An error is returned and SET is aborted if the value stored at the key is not a string
}

func (s *RedisStore) Set(key, value string, options SetOptions) {

	/*
		If GET was not specified, one of the following:
		Null bulk string reply in the following two cases.
		The key doesn’t exist and XX/IFEQ/IFDEQ was specified. The key was not created.
		The key exists, and NX was specified or a specified IFEQ/IFNE/IFDEQ/IFDNE condition is false. The key was not set.
		Simple string reply: OK: The key was set.
		If GET was specified, one of the following:
		Null bulk string reply: The key didn't exist before the SET operation, whether the key was created of not.
		Bulk string reply: The previous value of the key, whether the key was set or not.

	*/

	s.mu.Lock()
	defer s.mu.Unlock()

	s.db[key] = RedisObject{Type: StringObject, Value: value}
}

func (s *RedisStore) Get(key string) (string, bool, error) {
	s.mu.RLock()

	defer s.mu.RUnlock()

	va, exists := s.db[key]

	if !exists {
		return "", false, nil
	}

	if va.Type != StringObject {
		return "", false, ErrWrongType
	}

	if time.Now().After(va.ExpiresAt) {
		delete(s.db, key)
		return "", false, nil
	}

	value, ok := va.Value.(string)

	if !ok {
		return "", false, ErrWrongType
	}

	return value, true, nil

}

func (s *RedisStore) Incr(key string) (int, error) {
	// Appends a value to the string stored at key. If key does not exist, it is created with an empty string, so in that case, APPEND behaves like SET.

	s.mu.Lock()
	defer s.mu.Unlock()

	obj, exists := s.db[key]

	if !exists {
		s.db[key] = RedisObject{
			Type:  StringObject,
			Value: "1",
		}

		return 1, nil
	}

	strVal, ok := obj.Value.(string)

	if !ok || obj.Type != StringObject {
		return 0, ErrWrongType
	}

	n, err := strconv.Atoi(strVal)

	if err != nil {
		return 0, err
	}

	n++

	obj.Value = strconv.Itoa(n)

	s.db[key] = obj

	return n, nil

}
