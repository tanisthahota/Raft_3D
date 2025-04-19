package main

import (
	"sync"
)

// KVStore is a simple in-memory key-value store.
type KVStore struct {
	mu   sync.RWMutex
	data map[string]string
}

// NewKVStore creates a new KVStore.
func NewKVStore() *KVStore {
	return &KVStore{
		data: make(map[string]string),
	}
}

// Get returns the value for the given key.
func (kvs *KVStore) Get(key string) (string, bool) {
	kvs.mu.RLock()
	defer kvs.mu.RUnlock()
	val, ok := kvs.data[key]
	return val, ok
}

// Set sets the value for the given key.
func (kvs *KVStore) Set(key, value string) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	kvs.data[key] = value
}

// Delete deletes the value for the given key.
func (kvs *KVStore) Delete(key string) {
	kvs.mu.Lock()
	defer kvs.mu.Unlock()
	delete(kvs.data, key)
}

//hi cutie
//pls reply
//pls reply

