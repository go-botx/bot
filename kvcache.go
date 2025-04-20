package bot

import (
	"sync"
	"time"
)

type kvCache[K comparable, V any] struct {
	entries            map[K]V
	entriesExpirations map[K]time.Time
	validTime          time.Duration
	mu                 sync.RWMutex
}

func newKVCache[K comparable, V any](validTime time.Duration) *kvCache[K, V] {
	return &kvCache[K, V]{
		entries:            make(map[K]V),
		entriesExpirations: make(map[K]time.Time),
	}
}

func (uc *kvCache[K, V]) Set(k K, v V) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.entries[k] = v
	uc.entriesExpirations[k] = time.Now().Add(uc.validTime)
}

func (uc *kvCache[K, V]) Get(k K) (v V, ok bool) {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	v, ok = uc.entries[k]
	if ok {
		exp, okExp := uc.entriesExpirations[k]
		if !okExp || exp.Before(time.Now()) {
			ok = false
		}
	}
	return
}

func (uc *kvCache[K, V]) ReplaceAll(entries map[K]V) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.entries = make(map[K]V, len(entries))
	exp := time.Now().Add(uc.validTime)
	for k, v := range entries {
		uc.entries[k] = v
		uc.entriesExpirations[k] = exp
	}
}
