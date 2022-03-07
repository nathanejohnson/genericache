// Package genericache is a simple thread safe cache with no evictions, no expiration.
package genericache

import "sync"

// Filler is a function that is meant to fetch values on cache miss.
// An error returned from this function will be returned from the GeneriCache.Get() method.
type Filler[K comparable, V any] func(key K) (V, error)

// GeneriCache implements a simple cache with no eviction.
type GeneriCache[K comparable, V any] struct {
	mu          sync.Mutex
	entries     map[K]*cacheEntry[V]
	fill        Filler[K, V]
	retryErrors bool
}

type cacheEntry[V any] struct {
	sync.Mutex
	v   V
	err error
}

// NewGeneriCache - factory method for GeneriCache.  Use this to initialize.
// fillFunc is a method meant to fill the cache on cache miss.  retryErrors indicates
// whether an error received from fillFunc should indicate to try the filler function
// on next Get() operation.  If this is false, it will return the same error over and
// over and not call fillFunc more than once per key.  If true, it will call fillFunc on subsequent
// Get() calls until error is nil.
func NewGeneriCache[K comparable, V any](fillFunc Filler[K, V], retryErrors bool) *GeneriCache[K, V] {
	return &GeneriCache[K, V]{
		entries:     make(map[K]*cacheEntry[V]),
		retryErrors: retryErrors,
		fill:        fillFunc,
	}
}

// Get - Get a value from the cache.  On cache miss, it will call the Filler function.
func (gc *GeneriCache[K, V]) Get(key K) (V, error) {
	gc.mu.Lock()
	var ce *cacheEntry[V]
	var ok bool
	if ce, ok = gc.entries[key]; ok {
		// cache hit
		gc.mu.Unlock()
		// lock here to make sure a fill operation isn't in flight.
		ce.Lock()
		if ce.err == nil || !gc.retryErrors {
			defer ce.Unlock()
			return ce.v, ce.err
		}
	} else {
		// cache miss, create an entry to hold the value and lock it.
		ce = &cacheEntry[V]{}
		ce.Lock()
		gc.entries[key] = ce
		gc.mu.Unlock()
	}
	// if we get here, gc is unlocked, ce is locked.
	defer ce.Unlock()
	ce.v, ce.err = gc.fill(key)
	return ce.v, ce.err
}
