package cache

import (
	"container/list"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type item struct {
	Object     interface{}
	Expiration int64
}

func (i *item) Expired() bool {
	if i.Expiration < time.Now().UnixNano() {
		return true
	}
	return false
}

type cache struct {
	itemsMap          map[string]*list.Element
	itemsList         *list.List
	defaultExpiration time.Duration
	mu                sync.RWMutex
	maxSize           int64
	janitor           *janitor
}

type Cache struct {
	*cache
}

func (c *cache) Set(key string, val interface{}, expiration time.Duration) error {
	var exp int64
	if expiration == DefaultExpiration {
		expiration = c.defaultExpiration
	}
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}
	c.mu.Lock()
	oldVal, exist := c.itemsMap[key]
	if exist {
		c.itemsList.Remove(oldVal)
	}
	iter := c.itemsList.PushFront(item{Object: val, Expiration: exp})
	if iter != nil {
		c.itemsMap[key] = iter
	} else {
		c.mu.Unlock()
		return fmt.Errorf("set a new element err!")
	}
	if int64(c.itemsList.Len()) > c.maxSize {
		e := c.itemsList.Back()
		c.itemsList.Remove(e)
	}
	c.mu.Unlock()
	return nil
}

func (c *cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	elem, exist := c.itemsMap[key]
	if exist {
		c.itemsList.MoveToFront(elem)
		frontElem := c.itemsList.Front()
		val := frontElem.Value
		v, ok := val.(item)
		if ok {
			c.mu.RUnlock()
			return v.Object, true
		}
	}
	c.mu.RUnlock()
	return nil, false
}

func (c *cache) DeleteAllExpiredItems() {
	for k, v := range c.itemsMap {
		if it, ok := v.Value.(item); ok {
			if it.Expired() {
				c.mu.Lock()
				c.itemsList.Remove(v)
				delete(c.itemsMap, k)
				c.mu.Unlock()
			}
		}
	}
}

func NewCache(maxSize int64, d time.Duration) *Cache {
	itemsMap := make(map[string]*list.Element)
	itemsList := list.New()
	c := &cache{
		itemsMap:          itemsMap,
		itemsList:         itemsList,
		maxSize:           maxSize,
		defaultExpiration: 3600 * time.Second,
	}
	runJanitor(c, d)
	C := &Cache{c}
	runtime.SetFinalizer(C, stopJanitor)
	return C
}

type janitor struct {
	interval time.Duration
	stop     chan bool
}

func (j *janitor) run(c *cache) {
	ticker := time.NewTicker(j.interval)
	for {
		select {
		case <-ticker.C:
			c.DeleteAllExpiredItems()
		case <-j.stop:
			ticker.Stop()
			return
		}
	}
}

func runJanitor(c *cache, d time.Duration) {
	j := &janitor{
		interval: d,
		stop:     make(chan bool),
	}
	c.janitor = j
	go j.run(c)
}

func stopJanitor(c *Cache) {
	c.janitor.stop <- true
}
