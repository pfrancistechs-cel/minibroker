package broker

import (
	"sync"

	"github.com/pfrancistechs-cel/minibroker/backend/internal/domain"
)

// Broadcaster manages real-time subscribers for topic event streams (SSE).
type Broadcaster struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan domain.Message]struct{}
}

// NewBroadcaster initializes an event broadcaster.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[string]map[chan domain.Message]struct{}),
	}
}

// Subscribe opens a listener channel for a topic.
func (b *Broadcaster) Subscribe(topic string) chan domain.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan domain.Message, 128)
	if _, ok := b.subscribers[topic]; !ok {
		b.subscribers[topic] = make(map[chan domain.Message]struct{})
	}
	b.subscribers[topic][ch] = struct{}{}
	return ch
}

// Unsubscribe removes a channel listener.
func (b *Broadcaster) Unsubscribe(topic string, ch chan domain.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[topic]; ok {
		delete(subs, ch)
		close(ch)
		if len(subs) == 0 {
			delete(b.subscribers, topic)
		}
	}
}

// Broadcast dispatches a message to all active listeners of a topic.
func (b *Broadcaster) Broadcast(topic string, msg domain.Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[topic]; ok {
		for ch := range subs {
			select {
			case ch <- msg:
			default:
				// Buffer full; avoid blocking broker
			}
		}
	}
}

// ActiveSubscribers returns total count of streaming connections.
func (b *Broadcaster) ActiveSubscribers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	count := 0
	for _, subs := range b.subscribers {
		count += len(subs)
	}
	return count
}
