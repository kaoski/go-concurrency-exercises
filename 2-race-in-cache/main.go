//////////////////////////////////////////////////////////////////////
//
// Given is some code to cache key-value pairs from a database into
// the main memory (to reduce access time). Note that golang's map are
// not entirely thread safe. Multiple readers are fine, but multiple
// writers are not. Change the code to make this thread safe.
//

package main

import (
	"container/list"
	"fmt"
	"sync"
	"testing"
	"time"
)

// CacheSize determines how big the cache can grow
const CacheSize = 100
const SHARD = 19

// KeyStoreCacheLoader is an interface for the KeyStoreCache
type KeyStoreCacheLoader interface {
	// Load implements a function where the cache should gets it's content from
	Load(string) string
}

type page struct {
	Key   string
	Value string
}

// KeyStoreCache is a LRU cache for string key-value pairs
type KeyStoreCache struct {
	cache map[string]*list.Element
	pages list.List
	load  func(string) string
	lock  *sync.Mutex
}

type Cache struct {
	cacheMap map[int]*KeyStoreCache
}

func (c *Cache) len() (int, int) {
	cacheLength := 0
	pagesLength := 0
	for _, item := range c.cacheMap {
		cacheLength += len(item.cache)
		pagesLength += item.pages.Len()
	}
	return cacheLength, pagesLength
}

func hash(key string) int {
	sum := 0
	for _, r := range []byte(key) {
		sum += int(r)
	}
	return sum % SHARD
}

// New creates a new KeyStoreCache
func New(load KeyStoreCacheLoader) *Cache {
	cacheMap := make(map[int]*KeyStoreCache)
	for i := 0; i < SHARD; i++ {
		cacheMap[i] = &KeyStoreCache{
			load:  load.Load,
			cache: make(map[string]*list.Element),
			lock:  &sync.Mutex{},
		}
	}
	return &Cache{cacheMap: cacheMap}
}

func (c *Cache) Get(key string) string {
	hashValue := hash(key)
	keyStoreCache := c.cacheMap[hashValue]
	return c.cacheMap[hashValue].Get(key, keyStoreCache.lock)
}

// Get gets the key from cache, loads it from the source if needed
func (k *KeyStoreCache) Get(key string, lock *sync.Mutex) string {
	lock.Lock()
	defer lock.Unlock()
	if e, ok := k.cache[key]; ok {
		k.pages.MoveToFront(e)
		return e.Value.(page).Value
	}
	// Miss - load from database and save it in cache
	p := page{key, k.load(key)}
	// if cache is full remove the least used item
	if len(k.cache) >= CacheSize {
		end := k.pages.Back()
		// remove from map
		delete(k.cache, end.Value.(page).Key)
		// remove from list
		k.pages.Remove(end)
	}
	k.pages.PushFront(p)
	k.cache[key] = k.pages.Front()
	return p.Value
}

// Loader implements KeyStoreLoader
type Loader struct {
	DB *MockDB
}

// Load gets the data from the database
func (l *Loader) Load(key string) string {
	val, err := l.DB.Get(key)
	if err != nil {
		panic(err)
	}

	return val
}

func run(t *testing.T) (*Cache, *MockDB) {
	loader := Loader{
		DB: GetMockDB(),
	}
	cache := New(&loader)

	RunMockServer(cache, t)

	return cache, loader.DB
}

func main() {
	start := time.Now()
	run(nil)
	fmt.Println("Time Taken", time.Now().Sub(start))
}
