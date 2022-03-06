package genericache

import "sync"

type Filler[K comparable, V any] func(K) (V, error)

type GeneriCache[K comparable, V any] struct {
	sync.Mutex
	entries     map[K]*cacheEntry[V]
	fill        Filler[K, V]
	retryErrors bool
}

type cacheEntry[V any] struct {
	sync.Mutex
	v   V
	err error
}

func NewGeneriCache[K comparable, V any](f Filler[K, V], retryErrors bool) *GeneriCache[K, V] {
	return &GeneriCache[K, V]{
		entries:     make(map[K]*cacheEntry[V]),
		retryErrors: retryErrors,
		fill:        f,
	}
}

func (gc *GeneriCache[K, V]) Get(k K) (V, error) {
	gc.Lock()
	var ce *cacheEntry[V]
	var ok bool
	if ce, ok = gc.entries[k]; ok {
		// cache hit
		gc.Unlock()
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
		gc.entries[k] = ce
		gc.Unlock()
	}
	// if we get here, gc is unlocked, ce is locked.
	defer ce.Unlock()
	ce.v, ce.err = gc.fill(k)
	return ce.v, ce.err
}
