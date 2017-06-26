package cache

import ()

// HashFunc defines a hash function.
type HashFunc func(key interface{}) uint32

// ComboLRUCache internally uses multiple LRC caches to offer
// better concurrency. It simply uses a hash function to
// route a key to a specific LRC cache.
// ComboLRCCache is thread-safe.
type ComboLRUCache struct {
	hash HashFunc
	//buckets uint32
	caches []Cache
}

// NewComboLRUCache creates a ComboLRUCache instance.
func NewComboLRUCache(sz int, bs int, h HashFunc) Cache {
	if bs < 1 {
		bs = 1
	}
	if sz < bs {
		sz = bs
	}
	combo := &ComboLRUCache{
		hash: h,
		//buckets: uint32(bs),
		caches: make([]Cache, bs),
	}
	for i := range combo.caches {
		combo.caches[i] = NewLRUCache(sz / bs)
	}
	return combo
}

// Get returns the value identifiedy by key. If not found,
// an error is returned.
func (combo *ComboLRUCache) Get(key interface{}) (res interface{}, err error) {
	c := combo.routeKey(key)
	return c.Get(key)
}

// Set adds or updates the key-value pair to or in the cache.
func (combo *ComboLRUCache) Set(key, val interface{}) error {
	c := combo.routeKey(key)
	return c.Set(key, val)
}

// routeKey chooses a LRC cache instance by the hash function.
func (combo *ComboLRUCache) routeKey(key interface{}) Cache {
	h := combo.hash(key)
	idx := h % uint32(len(combo.caches))
	return combo.caches[idx]
}