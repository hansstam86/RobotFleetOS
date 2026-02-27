package state

import (
	"context"
	"sync"
)

// MemoryStore is an in-memory implementation of Store for development and testing.
type MemoryStore struct {
	mu    sync.RWMutex
	data  map[string][]byte
	watch map[string][]chan WatchEvent
}

// NewMemoryStore returns a new in-memory key-value store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:  make(map[string][]byte),
		watch: make(map[string][]chan WatchEvent),
	}
}

func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	if !ok {
		return nil, nil
	}
	cp := make([]byte, len(v))
	copy(cp, v)
	return cp, nil
}

func (s *MemoryStore) Put(ctx context.Context, key string, value []byte) error {
	s.mu.Lock()
	prev := s.data[key]
	s.data[key] = append([]byte(nil), value...)
	s.notifyLocked(key, value, false)
	s.mu.Unlock()
	_ = prev
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	delete(s.data, key)
	s.notifyLocked(key, nil, true)
	s.mu.Unlock()
	return nil
}

func (s *MemoryStore) notifyLocked(key string, value []byte, delete bool) {
	ev := WatchEvent{Key: key, Value: value, Delete: delete}
	for prefix, chans := range s.watch {
		if prefix == "" || (len(key) >= len(prefix) && key[0:len(prefix)] == prefix) {
			for _, ch := range chans {
				select {
				case ch <- ev:
				default:
					// non-blocking; skip if channel full
				}
			}
		}
	}
}

func (s *MemoryStore) Watch(ctx context.Context, prefix string) (<-chan WatchEvent, error) {
	ch := make(chan WatchEvent, 16)
	s.mu.Lock()
	s.watch[prefix] = append(s.watch[prefix], ch)
	s.mu.Unlock()
	go func() {
		<-ctx.Done()
		s.mu.Lock()
		defer s.mu.Unlock()
		for i, c := range s.watch[prefix] {
			if c == ch {
				s.watch[prefix] = append(s.watch[prefix][:i], s.watch[prefix][i+1:]...)
				break
			}
		}
		close(ch)
	}()
	return ch, nil
}
