package lru

import "container/list"

type Value interface{
	Len() int
}

type entry struct{
	key string
	value Value
}

type Cache struct {
	maxByte int64 // set the max size of storage space can be used
	list    *list.List
	cache	map[string]*list.Element
	OnEvicted func(key string, value Value)
}

func New(maxByte int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxByte: maxByte,
		list: list.New(),
		cache: make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
} 

func (c *Cache) Get(key string) (Value, bool){
	if listEle, ok := c.cache[key]; ok{
		// listEle is the most recent used, 
		// move to the back to make it least possible to be purged
		c.list.MoveToBack(listEle)
		
		kv, ok := listEle.Value.(*entry)
		if ok {
			return kv.value, true
		}
	}

	return nil, false
}
