// Package cache provides a simple implementation of LRC cache.
package cache

import (
	"container/list"
	"errors"
	"fmt"
	"io"
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
	Items() int
}

type item struct {
	key, val interface{}
	ts       time.Time
}

// LRUCache implements a cache using LRC algorithm.
// It is thread-safe.
type LRUCache struct {
	sz    int
	count int

	expiry time.Duration

	// store holds the key-value pairs.
	store map[interface{}]*list.Element

	// l maintains the items in the order of last-accessed time,
	// with the front being the latest and the end beging the oldest.
	l *list.List

	sync.Mutex
}

//NewLRUCache creates a LRCCache instance.
func NewLRUCache(sz int, expiry time.Duration) Cache {
	if sz <= 0 {
		log.Fatal("Size must be greater than 0")
	}

	c := &LRUCache{
		sz:     sz,
		expiry: expiry,
		store:  make(map[interface{}]*list.Element),
		count:  0,
		l:      list.New(),
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
	itm := elm.Value.(*item)
	val := itm.val
	if c.expiry > 0 && time.Since(itm.ts) > c.expiry {
		c.removeItem(elm)
	} else {
		c.updateItem(elm, val)
	}
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
		c.addItem(key, val)
		return nil
	}

	c.updateItem(elm, val)
	return nil
}

func (c *LRUCache) Items() int {
	if c == nil {
		return 0
	}

	return c.l.Len()
}

// updateItem updates the value of the item in elm,
// and move it to the front of the list.
// It must be called when the global lock is acquired.
func (c *LRUCache) updateItem(elm *list.Element, val interface{}) {
	itm := elm.Value.(*item)
	itm.val = val
	c.l.MoveToFront(elm)
}

// addItem adds a new item to the front of the list, and updates the counter.
// If the list is full, the last (oldest) item is removed.
// It must be called when the global lock is acquired.
func (c *LRUCache) addItem(key, val interface{}) (added *list.Element) {
	itm := &item{
		key: key,
		val: val,
		ts:  time.Now(),
	}

	if c.count >= c.sz {
		// Need to remove last item
		last := c.l.Back()
		c.removeItem(last)
	}

	added = c.l.PushFront(itm)
	c.store[key] = added
	c.count++
	return
}

// removeItem removes an item from the map and list.
// It must be called when the global lock is acquired.
func (c *LRUCache) removeItem(elm *list.Element) {
	if elm == nil {
		return
	}
	delete(c.store, elm.Value.(*item).key)
	c.l.Remove(elm)
	c.count--
}

func (c *LRUCache) PrintAll(w io.Writer, sep string) {
	c.Lock()
	defer c.Unlock()
	for e := c.l.Front(); e != nil; e = e.Next() {
		fmt.Fprintf(w, "%v%s", e.Value.(*item).val, sep)
	}
}
