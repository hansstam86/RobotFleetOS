package messaging

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NewBusFromURL returns a Bus. If url is empty or "memory", returns an in-memory bus.
// Otherwise treats url as NATS server (e.g. "nats://localhost:4222") and returns a NATS bus.
func NewBusFromURL(url string) (Bus, error) {
	url = strings.TrimSpace(url)
	if url == "" || url == "memory" {
		return NewMemoryBus(), nil
	}
	return NewNATSBus(url)
}

const keyHeader = "X-Key"

// NATSBus implements Bus using NATS for multi-process deployment.
type NATSBus struct {
	nc     *nats.Conn
	subs   []*nats.Subscription
	subsMu sync.Mutex
}

// NewNATSBus connects to NATS at url (e.g. "nats://localhost:4222") and returns a Bus.
// Retries with backoff so containers can start before NATS is ready; reconnects automatically if NATS restarts.
func NewNATSBus(url string) (*NATSBus, error) {
	opts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
	}
	var nc *nats.Conn
	var err error
	for attempt := 0; attempt < 30; attempt++ {
		nc, err = nats.Connect(url, opts...)
		if err == nil {
			break
		}
		time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	return &NATSBus{nc: nc}, nil
}

// Close closes the NATS connection.
func (b *NATSBus) Close() {
	b.nc.Drain()
}

// Publish publishes a message to the topic with optional key in header.
func (b *NATSBus) Publish(ctx context.Context, topic string, key string, value []byte) error {
	msg := nats.NewMsg(topic)
	msg.Data = value
	if key != "" {
		msg.Header.Set(keyHeader, key)
	}
	return b.nc.PublishMsg(msg)
}

// Subscribe registers a handler for the topic. Messages are delivered asynchronously.
// When ctx is cancelled, the subscription is unsubscribed.
func (b *NATSBus) Subscribe(ctx context.Context, topic string, handler func(key string, value []byte) error) error {
	sub, err := b.nc.Subscribe(topic, func(m *nats.Msg) {
		key := m.Header.Get(keyHeader)
		_ = handler(key, m.Data)
	})
	if err != nil {
		return err
	}
	b.subsMu.Lock()
	b.subs = append(b.subs, sub)
	b.subsMu.Unlock()
	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
	}()
	return nil
}
