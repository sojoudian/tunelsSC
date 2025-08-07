// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// LRU cache implementation - now using generics internally
package cache

import (
	"encoding/gob"
	"io"
	"os"
)

// Value interface that cache values must implement
type Value interface {
	Size() int
}

// Item represents a key-value pair (for backward compatibility)
type Item struct {
	Key   string
	Value Value
}

// LRUCache maintains backward compatibility by wrapping the generic implementation
type LRUCache struct {
	*LRUCacheGeneric[string, Value]
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(capacity uint64) *LRUCache {
	return &LRUCache{
		LRUCacheGeneric: NewLRUCacheGeneric[string, Value](capacity),
	}
}

// SaveItems saves cache items to a writer
func (lru *LRUCache) SaveItems(w io.Writer) error {
	items := lru.Items()
	// Convert generic items to legacy items
	legacyItems := make([]Item, len(items))
	for i, item := range items {
		legacyItems[i] = Item{Key: item.Key, Value: item.Value}
	}
	encoder := gob.NewEncoder(w)
	return encoder.Encode(legacyItems)
}

// SaveItemsToFile saves cache items to a file
func (lru *LRUCache) SaveItemsToFile(path string) error {
	if wr, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return err
	} else {
		defer wr.Close()
		return lru.SaveItems(wr)
	}
}

// LoadItems loads cache items from a reader
func (lru *LRUCache) LoadItems(r io.Reader) error {
	items := make([]Item, 0)
	decoder := gob.NewDecoder(r)
	if err := decoder.Decode(&items); err != nil {
		return err
	}

	for _, item := range items {
		lru.Set(item.Key, item.Value)
	}

	return nil
}

// LoadItemsFromFile loads cache items from a file
func (lru *LRUCache) LoadItemsFromFile(path string) error {
	if rd, err := os.Open(path); err != nil {
		return err
	} else {
		defer rd.Close()
		return lru.LoadItems(rd)
	}
}

// Items returns all items in the cache (backward compatibility)
func (lru *LRUCache) Items() []Item {
	genericItems := lru.LRUCacheGeneric.Items()
	items := make([]Item, len(genericItems))
	for i, item := range genericItems {
		items[i] = Item{Key: item.Key, Value: item.Value}
	}
	return items
}
