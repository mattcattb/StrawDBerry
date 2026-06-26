package store

func (rs *RedisStore) Del() {

}

func (rs *RedisStore) Ttl() {

}

func (rs *RedisStore) Info() {

}

func (rs *RedisStore) Expire() {

}

func (rs *RedisStore) Exists(keys []string) int {

	rs.mu.RLock()
	defer rs.mu.RUnlock()

	c := 0
	for _, key := range keys {
		_, exists := rs.db[key]
		if exists {
			c++
		}
	}

	return c
}
