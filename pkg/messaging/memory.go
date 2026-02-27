package messaging

import (
	"context"
	"sync"
)

// MemoryBus is an in-memory implementation of Bus for development and testing.
// Topics are keyed by topic name; each topic has a list of subscribers that receive copies of messages.
type MemoryBus struct {
	mu          sync.RWMutex
	subscribers map[string][]func(key string, value []byte) error
}

// NewMemoryBus returns a new in-memory message bus.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{
		subscribers: make(map[string][]func(key string, value []byte) error),
	}
}

// Publish sends a message to all subscribers of the topic.
func (b *MemoryBus) Publish(ctx context.Context, topic string, key string, value []byte) error {
	b.mu.RLock()
	handlers := b.subscribers[topic]
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(key, value); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler for the topic. All messages to the topic are delivered to the handler.
func (b *MemoryBus) Subscribe(ctx context.Context, topic string, handler func(key string, value []byte) error) error {
	b.mu.Lock()
	b.subscribers[topic] = append(b.subscribers[topic], handler)
	b.mu.Unlock()
	return nil
}
