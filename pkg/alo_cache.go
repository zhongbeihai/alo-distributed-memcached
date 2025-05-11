package pkg

import (
	"fmt"
	"log"
	"sync"

	singleflight "github.com/alo-distributed-memcached/pkg/single_flight"
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
	// HTTPPool implement PeerPicker interface. When the data is not in current node, current node will use HTTPPool.PickPeer()
	// to get the **HTTPGetter** of other node (not the other node) that has the data.
	peerPicker PeerPicker
	loader *singleflight.CallsGroup
}

var (
	mu          sync.RWMutex
	globeGroups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: ConcurrentCache{cacheSize: cacheBytes},
		loader: &singleflight.CallsGroup{},
	}
	globeGroups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := globeGroups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.Get(key); ok {
		log.Println("Cache hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) RegisterPeerPicker(peerPicker PeerPicker) {
	if g.peerPicker != nil {
		panic("RegisterPeerPick() called more than once")
	}
	g.peerPicker = peerPicker
}

func (g *Group) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peerPicker != nil {
			if peer, ok := g.peerPicker.PickPeer(key); ok{
				if value, err = g.getFromPeer(peer, key); err == nil{
					return value, nil
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}

	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.GetDataFromPeer(g.name, key)
	if err != nil {
		return ByteView{}, err
	}

	return ByteView{bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)

	return val, nil
}

func (g *Group) populateCache(key string, val ByteView) {
	g.mainCache.Add(key, val)
}
