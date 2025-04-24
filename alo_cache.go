package alodistributedmencached

import (
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache ConcurrentCache
}

var (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group{
	if getter == nil{
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name: name,
		getter: getter,
		mainCache: ConcurrentCache{cacheSize: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group{
	mu.RLock();
	defer mu.Unlock()
	g := groups[name]
	mu.Unlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error){
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok{
		log.Println("Cache hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error){
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error){
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)

	return val, nil
}

func (g *Group) populateCache(key string, val ByteView){
	g.mainCache.Add(key, val)
}