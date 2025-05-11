package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunction func(data []byte) uint32

type ConsistentHashMap struct {
	hashFunc     HashFunction
	replicas     int
	nodeHashKeys []int          // store all virtual node hash
	hashMap      map[int]string // map virtual node to actual node
}

func NewConsistentHashMap(replicas int, hashFunc HashFunction) *ConsistentHashMap {
	m := &ConsistentHashMap{
		replicas:     replicas,
		nodeHashKeys: make([]int, 0),
		hashMap:      make(map[int]string, 0),
		hashFunc:     hashFunc,
	}
	if hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

func (c *ConsistentHashMap) AddNode(actualNodeKeys ...string) {
	for _, key := range actualNodeKeys {
		for i := 0; i < c.replicas; i++ {
			nodeHash := int(c.hashFunc([]byte(strconv.Itoa(i) + key)))
			c.hashMap[nodeHash] = key
			c.nodeHashKeys = append(c.nodeHashKeys, nodeHash)
		}
	}
	sort.Ints(c.nodeHashKeys)
}

func (c *ConsistentHashMap) GetNode(key string) string {
	if len(c.nodeHashKeys) == 0 {
		return ""
	}

	hash := int(c.hashFunc([]byte(key)))
	idx := sort.Search(len(c.nodeHashKeys), func(i int) bool {
		return c.nodeHashKeys[i] >= hash
	})

	return c.hashMap[c.nodeHashKeys[idx%len(c.nodeHashKeys)]]
}
