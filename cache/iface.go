package cache

import (
	"fmt"
	"time"
)

const DefaultTimeout = 30

type item struct {
	Value    interface{} `json:"value"`
	Lifetime int64       `json:"lifetime"`
}

type Cache interface {
	Add(key string, value interface{}, timeout time.Duration) // add key-value pair to cache
	Delete(key string)                                        // delete key-value pair from cache
	Get(key string) interface{}                               // fetch value from given key in the cache, return the corrsponding value if not expired, nil otherwise
	cleanExpired()
}

func (i *item) String() string {
	return fmt.Sprintf("%s, expired at %v", i.Value, time.Unix(i.Lifetime, 0))
}
