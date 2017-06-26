// Package cache provides a simple implementation of LRC cache.
package cache

import (
	"container/list"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	// ErrCacheNotInit is returned when any Cache's method is called on nil.
	ErrCacheNotInit = errors.New("Cache is not init")
	// ErrItemNotFound is returned when the item is not in the cache.
	ErrItemNotFound = errors.New("Item not found in cache")
)

// Cache defines the interface for a cache object.
type Cache interface {
	Get(key interface{}) (res interface{}, err error)
	Set(key, val interface{}) error
}

type item struct {
	val interface{}
	ts  time.Time
}

// LRUCache implements a cache using LRC algorithm.
// It is thread-safe.
type LRUCache struct {
	sz    int
	count int

	// store holds the key-value pairs.
	store map[interface{}]*list.Element

	// l maintains the items in the order of last-accessed time,
	// with the front being the latest and the end beging the oldest.
	l *list.List

	sync.Mutex
}

//NewLRUCache creates a LRCCache instance.
func NewLRUCache(sz int) Cache {
	if sz <= 0 {
		log.Fatal("Size must be greater than 0")
	}

	c := &LRUCache{
		sz:    sz,
		store: make(map[interface{}]*list.Element),
		count: 0,
		l:     list.New(),
	}
	return c
}

// Get returns the value identifiedy by key. If not found,
// an error is returned.
func (c *LRUCache) Get(key interface{}) (interface{}, error) {
	if c == nil {
		return nil, ErrCacheNotInit
	}

	c.Lock()
	defer c.Unlock()

	elm, ok := c.store[key]
	if !ok {
		return nil, ErrItemNotFound
	}
	val := elm.Value.(*item).val
	c.updateItem(elm, val)
	return val, nil
}

// Set adds or updates the key-value pair to or in the cache.
func (c *LRUCache) Set(key, val interface{}) error {
	if c == nil {
		return ErrCacheNotInit
	}

	c.Lock()
	defer c.Unlock()

	elm, ok := c.store[key]
	if !ok {
		elm := c.addItem(key, val)
		c.store[key] = elm
		return nil
	}

	c.updateItem(elm, val)
	return nil
}

// updateItem updates the timestamp and value of the item in elm,
// and move it to the front of the list.
// It must be called when the global lock is acquired.
func (c *LRUCache) updateItem(elm *list.Element, val interface{}) {
	itm := elm.Value.(*item)
	itm.val = val
	itm.ts = time.Now()
	c.l.MoveToFront(elm)
}

// addItem adds a new item to the front of the list, and updates the counter.
// If the list is full, the last (oldest) item is removed.
// It must be called when the global lock is acquired.
func (c *LRUCache) addItem(key, val interface{}) *list.Element {
	itm := &item{
		val: val,
		ts:  time.Now(),
	}

	if c.count >= c.sz {
		// Need to remove last item
		last := c.l.Back()
		if last != nil {
			c.l.Remove(last)
			c.count--
		}
	}

	elm := c.l.PushFront(itm)
	c.count++
	return elm
}
