package store

import (
	"sync"
	"time"
)

type HashStore struct {
	mu     sync.RWMutex
	hashes map[string]map[string]string
	expiry map[string]time.Time
}
