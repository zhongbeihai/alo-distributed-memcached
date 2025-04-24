package alodistributedmencached

import (
	"sync"

	"github.com/alo-distributed-memcached/lru"
)

type Getter interface{
	Get(key string) ([]byte, error)
}

type GetterFunc func (key string) ([]byte, error)

func (f *GetterFunc) Get(key string) ([]byte, error){
	return f(key)
}


type ConcurrentCache struct {
	mu sync.Mutex
	lruCache *lru.Cache
	cacheSize int64
}


func (c *ConcurrentCache) Add(key string, value ByteView){
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lruCache == nil {
		c.lruCache = lru.New(c.cacheSize, nil)
	}
	c.lruCache.Add(key, value)
}

func (c *ConcurrentCache) Get(key string) (ByteView, bool){
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lruCache == nil {
		return ByteView{}, false
	}

	resp, ok := c.lruCache.Get(key)
	if ok {
		return resp.(ByteView), true
	}

	return ByteView{}, false
}


