package cache

import (
	"runtime"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type MemStore struct {
	mapper sync.Map
	stop   chan bool // stop signal to stop gorotine that check and clean expired items
}

func NewMemCache() *MemStore {
	cache := &MemStore{stop: make(chan bool)}
	go gc(cache, cache.stop)
	runtime.SetFinalizer(cache, func(ms *MemStore) {
		ms.stop <- true
	})

	return cache
}

func (c *MemStore) Add(key string, value interface{}, timeout time.Duration) {
	expireTime := time.Now().Add(timeout).Unix()

	c.mapper.Store(key, &item{Value: value,
		Lifetime: expireTime})

	log.Debugf("cache add: %s, expired at %v", key, time.Unix(expireTime, 0))
}

func (c *MemStore) Delete(key string) {
	c.mapper.Delete(key)
}

func (c *MemStore) Get(key string) interface{} {
	if v, exist := c.mapper.Load(key); exist {
		log.Debugf("cache hit: %s", key)
		return v.(*item).Value
	}

	return nil
}

func (c *MemStore) cleanExpired() {
	// go through every item store in cache, delete it when expired
	c.mapper.Range(func(key, value interface{}) bool {
		k := key.(string)
		expireTime := value.(*item).Lifetime

		if expireTime <= unixNow() {
			log.Debugf("cache expired: %s", k)
			c.Delete(k)
		}

		return true
	})

}
