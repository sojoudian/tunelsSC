// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Generic LRU cache implementation using Go 1.24 generics
package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// LRUCacheGeneric is a generic LRU cache implementation
type LRUCacheGeneric[K comparable, V Value] struct {
	mu sync.Mutex

	// list & table of *entryGeneric objects
	list  *list.List
	table map[K]*list.Element

	// Our current size, in bytes
	size uint64

	// How many bytes we are limiting the cache to
	capacity uint64
}

type entryGeneric[K comparable, V Value] struct {
	key           K
	value         V
	size          int
	time_accessed time.Time
}

// ItemGeneric represents a key-value pair in the cache
type ItemGeneric[K comparable, V Value] struct {
	Key   K
	Value V
}

// NewLRUCacheGeneric creates a new generic LRU cache
func NewLRUCacheGeneric[K comparable, V Value](capacity uint64) *LRUCacheGeneric[K, V] {
	return &LRUCacheGeneric[K, V]{
		list:     list.New(),
		table:    make(map[K]*list.Element),
		capacity: capacity,
	}
}

// Get retrieves a value from the cache
func (lru *LRUCacheGeneric[K, V]) Get(key K) (v V, ok bool) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		var zero V
		return zero, false
	}
	lru.moveToFront(element)
	return element.Value.(*entryGeneric[K, V]).value, true
}

// Set adds or updates a value in the cache
func (lru *LRUCacheGeneric[K, V]) Set(key K, value V) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value)
	}
}

// SetIfAbsent adds a value only if the key doesn't exist
func (lru *LRUCacheGeneric[K, V]) SetIfAbsent(key K, value V) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	if element := lru.table[key]; element != nil {
		lru.moveToFront(element)
	} else {
		lru.addNew(key, value)
	}
}

// Delete removes a key from the cache
func (lru *LRUCacheGeneric[K, V]) Delete(key K) bool {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= uint64(element.Value.(*entryGeneric[K, V]).size)
	return true
}

// Clear removes all items from the cache
func (lru *LRUCacheGeneric[K, V]) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.list.Init()
	lru.table = make(map[K]*list.Element)
	lru.size = 0
}

// SetCapacity updates the capacity of the cache
func (lru *LRUCacheGeneric[K, V]) SetCapacity(capacity uint64) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.capacity = capacity
	lru.checkCapacity()
}

// Stats returns cache statistics
func (lru *LRUCacheGeneric[K, V]) Stats() (length, size, capacity uint64, oldest time.Time) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entryGeneric[K, V]).time_accessed
	}
	return uint64(lru.list.Len()), lru.size, lru.capacity, oldest
}

// StatsJSON returns cache statistics as JSON
func (lru *LRUCacheGeneric[K, V]) StatsJSON() string {
	if lru == nil {
		return "{}"
	}
	l, s, c, o := lru.Stats()
	return fmt.Sprintf("{\"Length\": %v, \"Size\": %v, \"Capacity\": %v, \"OldestAccess\": \"%v\"}", l, s, c, o)
}

// Keys returns all keys in the cache
func (lru *LRUCacheGeneric[K, V]) Keys() []K {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	keys := make([]K, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entryGeneric[K, V]).key)
	}
	return keys
}

// Items returns all items in the cache
func (lru *LRUCacheGeneric[K, V]) Items() []ItemGeneric[K, V] {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	items := make([]ItemGeneric[K, V], 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(*entryGeneric[K, V])
		items = append(items, ItemGeneric[K, V]{Key: v.key, Value: v.value})
	}
	return items
}

func (lru *LRUCacheGeneric[K, V]) updateInplace(element *list.Element, value V) {
	valueSize := value.Size()
	entry := element.Value.(*entryGeneric[K, V])
	sizeDiff := valueSize - entry.size
	entry.value = value
	entry.size = valueSize
	lru.size += uint64(sizeDiff)
	lru.moveToFront(element)
	lru.checkCapacity()
}

func (lru *LRUCacheGeneric[K, V]) moveToFront(element *list.Element) {
	lru.list.MoveToFront(element)
	element.Value.(*entryGeneric[K, V]).time_accessed = time.Now()
}

func (lru *LRUCacheGeneric[K, V]) addNew(key K, value V) {
	newEntry := &entryGeneric[K, V]{key, value, value.Size(), time.Now()}
	element := lru.list.PushFront(newEntry)
	lru.table[key] = element
	lru.size += uint64(newEntry.size)
	lru.checkCapacity()
}

func (lru *LRUCacheGeneric[K, V]) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entryGeneric[K, V])
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= uint64(delValue.size)
	}
}

// Type aliases for backward compatibility
type StringLRUCache = LRUCacheGeneric[string, Value]

// NewStringLRUCache creates a string-keyed LRU cache for backward compatibility
func NewStringLRUCache(capacity uint64) *StringLRUCache {
	return NewLRUCacheGeneric[string, Value](capacity)
}
