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
	curByte int64 // storage size currently used 
	list    *list.List
	cache	map[string]*list.Element
	OnEvicted func(key string, value Value)
}

func New(maxByte int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxByte: maxByte,
		curByte: 0,
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

func (c *Cache) RemovdeOldest() {
	listEle := c.list.Front()
	if listEle != nil{
		c.list.Remove(listEle)

		kv := listEle.Value.(*entry)
		delete(c.cache, kv.key)

		c.curByte -= int64(len(kv.key) + kv.value.Len())

		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.list.MoveToBack(ele)
		
		kv := ele.Value.(*entry)
		c.curByte += int64(value.Len() - kv.value.Len())
		kv.value = value
	}else {
		ele := c.list.PushBack(&entry{key: key, value: value})
		c.cache[key] = ele
		c.curByte += int64(value.Len() + len(key))
	}

	for c.curByte > c.maxByte && c.maxByte != 0{
		c.RemovdeOldest()
	}
}

func (c *Cache) Len() int {
	return c.list.Len()
}
