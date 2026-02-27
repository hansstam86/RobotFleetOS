// Package state defines interfaces for distributed state stores used by each layer.
package state

import "context"

// Store is a key-value store with watch support (e.g. etcd, Consul).
// Used for config, leader election, and bounded scope state per layer.
type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
	Watch(ctx context.Context, prefix string) (<-chan WatchEvent, error)
}

// WatchEvent is a single change notification.
type WatchEvent struct {
	Key   string
	Value []byte
	Delete bool
}
