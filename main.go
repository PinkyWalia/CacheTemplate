package main

import (
	"fmt"
	"sync"
	"time"
)

// Errors for cache
const (
	ErrKeyNotFound = "key not found"
	ErrKeyExists   = "key already exists"
	ErrExpired     = "key expired"
)

// CacheItem is a struct for cache item
type CacheItem struct {
	value      interface{}
	expiration int64
}

// Cache is a struct for cache, Cache is a simple in-memory cache.
type Cache struct {
	data map[string]CacheItem //data map[string]interface{}
	mx   sync.RWMutex
}

// Cacher is an interface for cache
type Cacher interface {
	Set(key string, value interface{}, ttl int64) error
	Get(key string) (interface{}, error)
	Has(key string) (bool, error)
}

// NewCache is a constructor for Cache
func NewCache(options ...func(*Cache)) *Cache {
	c := &Cache{
		data: make(map[string]CacheItem),
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Set is a method for setting key-value pair
// If key already exists, and it's not expired, return error
// If key already exists, but it's expired, set new value and return nil
// If key doesn't exist, set new value and return nil
// If ttl is 0, set value without expiration
func (c *Cache) Set(key string, value interface{}, ttl int64) error {
	c.mx.RLock()
	d, ok := c.data[key]
	c.mx.RUnlock()
	if ok {
		if d.expiration == 0 || d.expiration > time.Now().Unix() {
			return fmt.Errorf(ErrKeyExists)
		}
	}

	var expiration int64

	if ttl > 0 {
		expiration = time.Now().Unix() + ttl
	}

	c.mx.Lock()
	c.data[key] = CacheItem{
		value:      value,
		expiration: expiration,
	}
	c.mx.Unlock()
	return nil
}

// Get is a method for getting value by key
// If key doesn't exist, return error
// If key exists, but it's expired, return error and delete key
// If key exists and it's not expired, return value
func (c *Cache) Get(key string) (interface{}, error) {

	_, err := c.Has(key)
	if err != nil {
		return nil, err
	}

	// safe return?
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.data[key].value, nil
}

// Has is a method for checking if key exists.
// If key doesn't exist, return false.
// If key exists, but it's expired, return false and delete key.
// If key exists and it's not expired, return true.
func (c *Cache) Has(key string) (bool, error) {
	c.mx.RLock()
	d, ok := c.data[key]
	c.mx.RUnlock()
	if !ok {
		return false, fmt.Errorf(ErrKeyNotFound)
	}

	if d.expiration != 0 && d.expiration < time.Now().Unix() {
		c.mx.Lock()
		delete(c.data, key)
		c.mx.Unlock()
		return false, fmt.Errorf(ErrExpired)
	}

	return true, nil
}

// type testItem struct {
// 	key   string
// 	value interface{}
// 	ttl   int64
// }

// func main() {

// 	c := NewCache()

// 	testItems := []testItem{
// 		{"key0", "value0", 0},
// 		{"key1", "value1", 1},
// 		{"key2", "value2", 20},
// 		{"key3", "value3", 30},
// 		{"key4", "value4", 40},
// 		{"key5", "value5", 50},
// 		{"key6", "value6", 60},
// 		{"key7", "value7", 70000000},
// 	}

// 	for _, item := range testItems {
// 		err := c.Set(item.key, item.value, item.ttl)
// 		fmt.Printf("errormsg %s\n", err)
// 	}

// 	for _, item := range testItems {
// 		value, err := c.Get(item.key)
// 		fmt.Printf("errormsg %s\n", err)
// 		fmt.Printf("Running on port %s\n", value)
// 	}
// }
