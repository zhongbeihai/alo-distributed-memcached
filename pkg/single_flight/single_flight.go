package singleflight

import "sync"

/*
	A call stand for an executing request
*/
type call struct {
	wg sync.WaitGroup
	val interface{}
	err error
}

type CallsGroup struct{
	mu sync.Mutex
	mapping map[string]*call // <key, *call>
}

func (g *CallsGroup) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	if g.mapping == nil {
		g.mapping = make(map[string]*call, 0)
	}
	if c, ok := g.mapping[key]; ok{
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	c := &call{}
	c.wg.Add(1)
	g.mapping[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()
	c.wg.Done()

	g.mu.Lock()
	delete(g.mapping, key)
	g.mu.Unlock()

	return c.val, c.err
}